package activity

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
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
			// Apply overrides if specified
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
			projectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get project name for display
			project, err := client.GetProject(cmd.Context(), projectID)
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

			// Set limit
			if limit > 0 {
				opts.Limit = limit
			}

			// Get recordings (activity)
			recordings, err := client.ListRecordings(cmd.Context(), projectID, opts)
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
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table or json")
	cmd.Flags().IntVarP(&limit, "limit", "l", 25, "Limit number of items shown")

	return cmd
}

// parseSince parses various time formats into a time.Time
func parseSince(s string) (time.Time, error) {
	now := time.Now()

	// Try parsing duration shorthand (e.g., "24h", "7d", "2w")
	s = strings.TrimSpace(strings.ToLower(s))

	// Handle human-friendly durations
	if strings.HasSuffix(s, "h") {
		hours, err := parseDurationValue(strings.TrimSuffix(s, "h"))
		if err == nil {
			return now.Add(-time.Duration(hours) * time.Hour), nil
		}
	}
	if strings.HasSuffix(s, "d") {
		days, err := parseDurationValue(strings.TrimSuffix(s, "d"))
		if err == nil {
			return now.AddDate(0, 0, -days), nil
		}
	}
	if strings.HasSuffix(s, "w") {
		weeks, err := parseDurationValue(strings.TrimSuffix(s, "w"))
		if err == nil {
			return now.AddDate(0, 0, -weeks*7), nil
		}
	}

	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try parsing as date only
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	// Try parsing relative phrases
	switch s {
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
	ID            int64     `json:"id"`
	Type          string    `json:"type"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	Creator       string    `json:"creator"`
	CreatorEmail  string    `json:"creator_email,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	URL           string    `json:"url"`
	ParentTitle   string    `json:"parent_title,omitempty"`
	ParentType    string    `json:"parent_type,omitempty"`
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
		table.AddHeader("TYPE", "TITLE", "BY", "UPDATED")
	} else {
		table.AddHeader("ID", "TYPE", "TITLE", "STATUS", "BY", "CREATED", "UPDATED")
	}

	now := time.Now()

	// Add rows
	for _, r := range recordings {
		if !table.IsTTY() {
			table.AddField(fmt.Sprintf("%d", r.ID))
		}

		// Type with color
		typeLabel := formatRecordingType(r.Type)
		table.AddField(typeLabel, cs.Muted)

		// Title
		table.AddField(r.Title)

		if !table.IsTTY() {
			table.AddField(r.Status, cs.Muted)
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
		"Todo":            "todo",
		"Message":         "message",
		"Document":        "document",
		"Comment":         "comment",
		"Upload":          "upload",
		"Todolist":        "todolist",
		"Question":        "question",
		"Question::Answer": "answer",
		"Schedule::Entry": "event",
		"Vault":           "vault",
	}

	if label, ok := typeLabels[t]; ok {
		return label
	}
	return strings.ToLower(t)
}
