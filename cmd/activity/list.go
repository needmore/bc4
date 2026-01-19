package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	coretableprinter "github.com/needmore/bc4/internal/tableprinter"
	"github.com/needmore/bc4/internal/ui"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var (
		accountID     string
		projectID     string
		sinceStr      string
		recordingType string
		personStr     string
		formatStr     string
		limit         int
	)

	cmd := &cobra.Command{
		Use:     "list [project]",
		Short:   "List recent project activity",
		Long:    `List recent activity and changes across a Basecamp project.`,
		Aliases: []string{"ls"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse project argument if provided (could be URL or ID)
			if len(args) > 0 {
				if parser.IsBasecampURL(args[0]) {
					parsed, err := parser.ParseBasecampURL(args[0])
					if err != nil {
						return fmt.Errorf("invalid Basecamp URL: %w", err)
					}
					if parsed.ResourceType != parser.ResourceTypeProject {
						return fmt.Errorf("URL is not for a project: %s", args[0])
					}
					// Override factory with URL-provided values
					if parsed.AccountID > 0 {
						f = f.WithAccount(strconv.FormatInt(parsed.AccountID, 10))
					}
					if parsed.ProjectID > 0 {
						f = f.WithProject(strconv.FormatInt(parsed.ProjectID, 10))
					}
				} else {
					// Treat as project ID
					f = f.WithProject(args[0])
				}
			}

			// Apply flag overrides (flags take precedence over URL)
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get project name for display
			project, err := client.GetProject(cmd.Context(), resolvedProjectID)
			if err != nil {
				return err
			}

			// Build options
			opts := &api.ActivityListOptions{}

			// Parse since flag
			if sinceStr != "" {
				since, err := parseSince(sinceStr)
				if err != nil {
					return fmt.Errorf("invalid --since value: %w", err)
				}
				opts.Since = &since
			}

			// Parse type filter
			if recordingType != "" {
				opts.RecordingTypes = parseTypes(recordingType)
			}

			// Parse person filter
			if personStr != "" {
				personID, err := parsePersonIdentifier(client, cmd.Context(), personStr)
				if err != nil {
					return fmt.Errorf("invalid --person value: %w", err)
				}
				opts.PersonID = personID
			}

			// Set limit
			if limit > 0 {
				opts.Limit = limit
			}

			// Get recordings (activity)
			recordings, err := client.ListRecordings(cmd.Context(), resolvedProjectID, opts)
			if err != nil {
				return err
			}

			// Check output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			if format == ui.OutputFormatJSON {
				return outputActivityJSON(recordings, project.Name)
			}

			// Display activity
			if len(recordings) == 0 {
				fmt.Println("No recent activity found")
				return nil
			}

			return renderActivityTable(recordings, project.Name)
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&sinceStr, "since", "", "Show activity since time (e.g., '24h', '7d', '2024-01-01')")
	cmd.Flags().StringVarP(&recordingType, "type", "t", "", "Filter by type: todo, message, document, comment, upload")
	cmd.Flags().StringVar(&personStr, "person", "", "Filter by person (ID, name, or email)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table or json")
	cmd.Flags().IntVarP(&limit, "limit", "l", 25, "Limit number of items shown")

	return cmd
}

// parseSince parses various time formats into a time.Time
func parseSince(s string) (time.Time, error) {
	now := time.Now()

	// Trim spaces but preserve case for RFC3339 parsing
	s = strings.TrimSpace(s)
	sLower := strings.ToLower(s)

	// Handle human-friendly durations
	if strings.HasSuffix(sLower, "h") {
		hours, err := parseDurationValue(strings.TrimSuffix(sLower, "h"))
		if err == nil {
			return now.Add(-time.Duration(hours) * time.Hour), nil
		}
	}
	if strings.HasSuffix(sLower, "d") {
		days, err := parseDurationValue(strings.TrimSuffix(sLower, "d"))
		if err == nil {
			return now.AddDate(0, 0, -days), nil
		}
	}
	if strings.HasSuffix(sLower, "w") {
		weeks, err := parseDurationValue(strings.TrimSuffix(sLower, "w"))
		if err == nil {
			return now.AddDate(0, 0, -weeks*7), nil
		}
	}

	// Try parsing as RFC3339 (preserve original case)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try parsing as date only (preserve original case)
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	// Try parsing relative phrases (use lowercase)
	switch sLower {
	case "today":
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location()), nil
	case "yesterday":
		y, m, d := now.AddDate(0, 0, -1).Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location()), nil
	case "this week":
		// Go to start of this week (Sunday)
		daysToSunday := int(now.Weekday())
		return now.AddDate(0, 0, -daysToSunday), nil
	case "last week":
		daysToSunday := int(now.Weekday())
		return now.AddDate(0, 0, -daysToSunday-7), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// parseDurationValue parses an integer from a string
func parseDurationValue(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

// parseTypes parses the type filter into a slice of recording types
func parseTypes(s string) []string {
	types := strings.Split(s, ",")
	result := make([]string, 0, len(types))

	typeMap := map[string]string{
		"todo":     "Todo",
		"todos":    "Todo",
		"message":  "Message",
		"messages": "Message",
		"msg":      "Message",
		"document": "Document",
		"doc":      "Document",
		"docs":     "Document",
		"comment":  "Comment",
		"comments": "Comment",
		"upload":   "Upload",
		"uploads":  "Upload",
		"file":     "Upload",
		"files":    "Upload",
	}

	for _, t := range types {
		t = strings.TrimSpace(strings.ToLower(t))
		if mapped, ok := typeMap[t]; ok {
			result = append(result, mapped)
		} else {
			// Use as-is if not in map
			result = append(result, t)
		}
	}

	return result
}

// ActivityOutput represents the JSON output format
type ActivityOutput struct {
	Project  string           `json:"project"`
	Activity []ActivityRecord `json:"activity"`
}

// ActivityRecord represents a single activity item for JSON output
type ActivityRecord struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Creator      string    `json:"creator"`
	CreatorEmail string    `json:"creator_email,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	URL          string    `json:"url"`
	ParentTitle  string    `json:"parent_title,omitempty"`
	ParentType   string    `json:"parent_type,omitempty"`
}

func outputActivityJSON(recordings []api.Recording, projectName string) error {
	output := ActivityOutput{
		Project:  projectName,
		Activity: make([]ActivityRecord, 0, len(recordings)),
	}

	for _, r := range recordings {
		record := ActivityRecord{
			ID:        r.ID,
			Type:      r.Type,
			Title:     r.Title,
			Status:    r.Status,
			Creator:   r.Creator.Name,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			URL:       r.AppURL,
		}
		if r.Creator.EmailAddress != "" {
			record.CreatorEmail = r.Creator.EmailAddress
		}
		if r.Parent != nil {
			record.ParentTitle = r.Parent.Title
			record.ParentType = r.Parent.Type
		}
		output.Activity = append(output.Activity, record)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func renderActivityTable(recordings []api.Recording, projectName string) error {
	// Create table
	table := tableprinter.New(os.Stdout)
	cs := table.GetColorScheme()

	// Print project header
	fmt.Printf("PROJECT: %s\n\n", projectName)

	// Add headers dynamically based on TTY mode
	if table.IsTTY() {
		table.AddHeader("TYPE", "TITLE", "CONTEXT", "BY", "UPDATED")
	} else {
		table.AddHeader("ID", "TYPE", "TITLE", "STATUS", "PARENT_TYPE", "PARENT_TITLE", "BY", "CREATED", "UPDATED")
	}

	now := time.Now()

	// Add rows
	for _, r := range recordings {
		if !table.IsTTY() {
			table.AddField(fmt.Sprintf("%d", r.ID))
		}

		// Type with color and icon
		typeLabel, typeColor := formatRecordingTypeWithStyle(r.Type, cs)
		table.AddField(typeLabel, typeColor)

		// Title with truncation for long titles
		title := r.Title
		if table.IsTTY() && len(title) > 60 {
			title = title[:57] + "..."
		}
		table.AddField(title)

		if !table.IsTTY() {
			table.AddField(r.Status, cs.Muted)

			// Parent info for non-TTY
			if r.Parent != nil {
				table.AddField(r.Parent.Type, cs.Muted)
				table.AddField(r.Parent.Title, cs.Muted)
			} else {
				table.AddField("", cs.Muted)
				table.AddField("", cs.Muted)
			}
		} else {
			// Context column for TTY (shows parent if exists)
			if r.Parent != nil {
				contextLabel := fmt.Sprintf("in %s", r.Parent.Title)
				if len(contextLabel) > 40 {
					contextLabel = contextLabel[:37] + "..."
				}
				table.AddField(contextLabel, cs.Muted)
			} else {
				table.AddField("", cs.Muted)
			}
		}

		// Creator
		table.AddField(r.Creator.Name, cs.Muted)

		if !table.IsTTY() {
			table.AddField(r.CreatedAt.Format(time.RFC3339), cs.Muted)
		}

		// Updated time
		table.AddTimeField(now, r.UpdatedAt)
		table.EndRow()
	}

	// Render
	return table.Render()
}

// formatRecordingType formats the recording type for display
func formatRecordingType(t string) string {
	typeLabels := map[string]string{
		"Todo":             "todo",
		"Message":          "message",
		"Document":         "document",
		"Comment":          "comment",
		"Upload":           "upload",
		"Todolist":         "todolist",
		"Question":         "question",
		"Question::Answer": "answer",
		"Schedule::Entry":  "event",
		"Vault":            "vault",
		"Card":             "card",
		"Card::Table":      "card_table",
		"Campfire":         "campfire",
	}

	if label, ok := typeLabels[t]; ok {
		return label
	}
	return strings.ToLower(t)
}

// formatRecordingTypeWithStyle returns the formatted type label and color function
func formatRecordingTypeWithStyle(t string, cs *coretableprinter.ColorScheme) (string, func(string) string) {
	type typeInfo struct {
		label string
		icon  string
		color func(string) string
	}

	typeMap := map[string]typeInfo{
		"Todo":             {label: "todo", icon: "✓", color: cs.Green},
		"Message":          {label: "message", icon: "💬", color: cs.Cyan},
		"Document":         {label: "document", icon: "📄", color: cs.Gray},
		"Comment":          {label: "comment", icon: "💭", color: cs.Muted},
		"Upload":           {label: "upload", icon: "📎", color: cs.Gray},
		"Todolist":         {label: "list", icon: "📋", color: cs.Green},
		"Question":         {label: "question", icon: "❓", color: cs.Cyan},
		"Question::Answer": {label: "answer", icon: "✏️", color: cs.Gray},
		"Schedule::Entry":  {label: "event", icon: "📅", color: cs.Cyan},
		"Vault":            {label: "vault", icon: "📁", color: cs.Gray},
		"Card":             {label: "card", icon: "🎴", color: cs.Cyan},
		"Card::Table":      {label: "board", icon: "📊", color: cs.Gray},
		"Campfire":         {label: "chat", icon: "🔥", color: cs.Gray},
	}

	if info, ok := typeMap[t]; ok {
		return fmt.Sprintf("%s %s", info.icon, info.label), info.color
	}

	// Default formatting
	return strings.ToLower(t), cs.Muted
}

// parsePersonIdentifier parses a person identifier (ID, name, or email) into a person ID
func parsePersonIdentifier(client *api.ModularClient, ctx context.Context, identifier string) (int64, error) {
	// Try parsing as ID first
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		return id, nil
	}

	// Otherwise, fetch all people and search by name or email
	people, err := client.GetAllPeople(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list people: %w", err)
	}

	// Search for matching person by name or email
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	var matches []api.Person
	for _, person := range people {
		nameMatch := strings.Contains(strings.ToLower(person.Name), identifier)
		emailMatch := strings.Contains(strings.ToLower(person.EmailAddress), identifier)
		if nameMatch || emailMatch {
			matches = append(matches, person)
		}
	}

	if len(matches) == 0 {
		return 0, fmt.Errorf("no person found matching '%s'", identifier)
	}

	if len(matches) > 1 {
		return 0, fmt.Errorf("multiple people found matching '%s' - please be more specific or use person ID", identifier)
	}

	return matches[0].ID, nil
}
