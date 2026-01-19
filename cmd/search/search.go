package search

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

// NewSearchCmd creates a new search command
func NewSearchCmd(f *factory.Factory) *cobra.Command {
	var (
		resourceType string
		projectID    string
		accountID    string
		formatStr    string
		limit        int
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search across all resources",
		Long: `Search across all Basecamp resources including todos, messages, documents, and cards.

The search performs a cross-resource discovery query against the Basecamp API,
respecting rate limits and pagination.`,
		Example: `  bc4 search "quarterly report"          # Global search across all resources
  bc4 search --type todo "bug fix"        # Search only todos
  bc4 search --type message "announcement" # Search only messages
  bc4 search --project 12345 "deadline"   # Search within a specific project
  bc4 search --type card "feature"        # Search only cards
  bc4 search -t document "spec"           # Search only documents`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			// Handle project argument override from URL
			if projectID != "" {
				if parser.IsBasecampURL(projectID) {
					parsed, err := parser.ParseBasecampURL(projectID)
					if err != nil {
						return fmt.Errorf("invalid Basecamp URL: %w", err)
					}
					if parsed.AccountID > 0 {
						f = f.WithAccount(strconv.FormatInt(parsed.AccountID, 10))
					}
					if parsed.ProjectID > 0 {
						f = f.WithProject(strconv.FormatInt(parsed.ProjectID, 10))
					}
				} else {
					f = f.WithProject(projectID)
				}
			}

			// Apply account override if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Build search options
			opts := api.SearchOptions{
				Query: query,
				Limit: limit,
			}

			// Set project scope if specified
			if projectID != "" {
				resolvedProjectID, err := f.ProjectID()
				if err != nil {
					return err
				}
				opts.ProjectID = resolvedProjectID
			}

			// Parse resource type filter
			if resourceType != "" {
				types, err := parseResourceTypes(resourceType)
				if err != nil {
					return err
				}
				opts.Types = types
			}

			// Perform search
			results, err := client.Search().Search(cmd.Context(), opts)
			if err != nil {
				return err
			}

			// Check output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			if format == ui.OutputFormatJSON {
				return outputSearchJSON(results, query)
			}

			// Display results
			if len(results) == 0 {
				fmt.Printf("No results found for '%s'\n", query)
				return nil
			}

			return renderSearchResults(results, query)
		},
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Flags
	cmd.Flags().StringVarP(&resourceType, "type", "t", "", "Filter by resource type: todo, message, document, card (comma-separated for multiple)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Scope search to a specific project (ID or URL)")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table or json")
	cmd.Flags().IntVarP(&limit, "limit", "l", 50, "Maximum number of results to return")

	return cmd
}

// parseResourceTypes parses the type filter into a slice of valid recording types
func parseResourceTypes(s string) ([]string, error) {
	types := strings.Split(s, ",")
	result := make([]string, 0, len(types))

	for _, t := range types {
		t = strings.TrimSpace(strings.ToLower(t))
		if mapped, ok := api.ValidSearchTypes[t]; ok {
			result = append(result, mapped)
		} else {
			validTypes := make([]string, 0, len(api.ValidSearchTypes))
			for k := range api.ValidSearchTypes {
				validTypes = append(validTypes, k)
			}
			return nil, fmt.Errorf("invalid type '%s'. Valid types: %s", t, strings.Join(validTypes, ", "))
		}
	}

	return result, nil
}

// SearchOutput represents the JSON output format
type SearchOutput struct {
	Query   string         `json:"query"`
	Count   int            `json:"count"`
	Results []SearchRecord `json:"results"`
}

// SearchRecord represents a single search result for JSON output
type SearchRecord struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Project      string    `json:"project"`
	ProjectID    int64     `json:"project_id"`
	Creator      string    `json:"creator"`
	CreatorEmail string    `json:"creator_email,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	URL          string    `json:"url"`
	ParentTitle  string    `json:"parent_title,omitempty"`
	ParentType   string    `json:"parent_type,omitempty"`
}

func outputSearchJSON(results []api.SearchResult, query string) error {
	output := SearchOutput{
		Query:   query,
		Count:   len(results),
		Results: make([]SearchRecord, 0, len(results)),
	}

	for _, r := range results {
		record := SearchRecord{
			ID:        r.ID,
			Type:      r.Type,
			Title:     r.Title,
			Status:    r.Status,
			Project:   r.Bucket.Name,
			ProjectID: r.Bucket.ID,
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
		output.Results = append(output.Results, record)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func renderSearchResults(results []api.SearchResult, query string) error {
	// Create table
	table := tableprinter.New(os.Stdout)
	cs := table.GetColorScheme()

	// Print search header
	fmt.Printf("Results for '%s' (%d found)\n\n", query, len(results))

	// Add headers dynamically based on TTY mode
	if table.IsTTY() {
		table.AddHeader("TYPE", "TITLE", "PROJECT", "UPDATED")
	} else {
		table.AddHeader("ID", "TYPE", "TITLE", "STATUS", "PROJECT", "PROJECT_ID", "BY", "CREATED", "UPDATED", "URL")
	}

	now := time.Now()

	// Add rows
	for _, r := range results {
		if !table.IsTTY() {
			table.AddField(fmt.Sprintf("%d", r.ID))
		}

		// Type with color
		typeLabel := formatResourceType(r.Type)
		table.AddField(typeLabel, cs.Muted)

		// Title (truncate if necessary for TTY display)
		title := r.Title
		if table.IsTTY() && len(title) > 50 {
			title = title[:47] + "..."
		}
		table.AddField(title)

		if !table.IsTTY() {
			table.AddField(r.Status, cs.Muted)
		}

		// Project name
		projectName := r.Bucket.Name
		if table.IsTTY() && len(projectName) > 20 {
			projectName = projectName[:17] + "..."
		}
		table.AddField(projectName, cs.Muted)

		if !table.IsTTY() {
			table.AddField(fmt.Sprintf("%d", r.Bucket.ID), cs.Muted)
			table.AddField(r.Creator.Name, cs.Muted)
			table.AddField(r.CreatedAt.Format(time.RFC3339), cs.Muted)
		}

		// Updated time
		table.AddTimeField(now, r.UpdatedAt)

		if !table.IsTTY() {
			table.AddField(r.AppURL, cs.Muted)
		}

		table.EndRow()
	}

	// Render
	return table.Render()
}

// formatResourceType formats the resource type for display
func formatResourceType(t string) string {
	typeLabels := map[string]string{
		"Todo":     "todo",
		"Message":  "message",
		"Document": "document",
		"Card":     "card",
		"Comment":  "comment",
		"Upload":   "upload",
		"Todolist": "todolist",
	}

	if label, ok := typeLabels[t]; ok {
		return label
	}
	return strings.ToLower(t)
}
