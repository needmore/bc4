package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/utils"
)

func newViewCmd() *cobra.Command {
	var accountID string
	var projectID string
	var formatStr string
	var jsonFields string
	var webView bool
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view [todo-id]",
		Short: "View details of a specific todo",
		Long:  `View detailed information about a specific todo item including description, assignees, due date, and completion status.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if we have auth
			if cfg.ClientID == "" || cfg.ClientSecret == "" {
				return fmt.Errorf("not authenticated. Run 'bc4' to set up authentication")
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

			// Use specified account or default
			if accountID == "" {
				accountID = authClient.GetDefaultAccount()
			}

			if accountID == "" {
				return fmt.Errorf("no account specified and no default account set. Use --account or run 'bc4 account select'")
			}

			// Use specified project or default
			if projectID == "" {
				projectID = cfg.DefaultProject
				if projectID == "" && cfg.Accounts != nil {
					if acc, ok := cfg.Accounts[accountID]; ok {
						projectID = acc.DefaultProject
					}
				}
			}

			if projectID == "" {
				return fmt.Errorf("no project specified and no default project set. Use --project or run 'bc4 project select'")
			}

			// Get token
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get authentication token: %w", err)
			}

			// Parse todo ID
			todoID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid todo ID: %s", args[0])
			}

			// Create API client
			apiClient := api.NewClient(accountID, token.AccessToken)

			// Get the todo details
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Fetch the todo details using the generic Get method
			var todo api.Todo
			todoPath := fmt.Sprintf("/buckets/%s/todos/%d.json", projectID, todoID)

			// Note: The Get method doesn't take context, so we just ensure the timeout is set
			_ = ctx // Mark as used
			err = apiClient.Get(todoPath, &todo)
			if err != nil {
				return fmt.Errorf("failed to get todo: %w", err)
			}

			// Open in browser if requested
			if webView {
				// For now, just print the URL since OpenInBrowser is not implemented
				fmt.Printf("Todo URL: https://3.basecamp.com/%s/buckets/%s/todos/%d\n", accountID, projectID, todoID)
				return nil
			}

			// Handle JSON output
			if formatStr == "json" {
				var output interface{} = todo

				// If specific fields requested, filter the output
				if jsonFields != "" {
					fields := strings.Split(jsonFields, ",")
					filtered := make(map[string]interface{})

					// Convert todo to map for field selection
					todoJSON, _ := json.Marshal(todo)
					var todoMap map[string]interface{}
					json.Unmarshal(todoJSON, &todoMap)

					for _, field := range fields {
						field = strings.TrimSpace(field)
						if val, ok := todoMap[field]; ok {
							filtered[field] = val
						}
					}
					output = filtered
				}

				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(output)
			}

			// Prepare output for pager
			var buf bytes.Buffer

			// Title style
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
			labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8"))

			// Status indicator
			statusIcon := "○"
			statusColor := "2" // Green
			if todo.Completed {
				statusIcon = "✓"
				statusColor = "1" // Red
			}
			statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))

			// Display todo content with status
			fmt.Fprintf(&buf, "\n%s %s\n\n", statusStyle.Render(statusIcon), titleStyle.Render(todo.Content))

			// Show metadata
			fmt.Fprintf(&buf, "%s %d\n", labelStyle.Render("ID:"), todo.ID)
			statusText := "active"
			if todo.Completed {
				statusText = "completed"
			}
			fmt.Fprintf(&buf, "%s %s\n", labelStyle.Render("Status:"), statusText)

			if todo.DueOn != nil && *todo.DueOn != "" {
				// Parse and format due date
				if dueDate, err := time.Parse("2006-01-02", *todo.DueOn); err == nil {
					daysUntil := int(time.Until(dueDate).Hours() / 24)
					dueText := dueDate.Format("January 2, 2006")

					if daysUntil == 0 {
						dueText += " (today)"
					} else if daysUntil == 1 {
						dueText += " (tomorrow)"
					} else if daysUntil > 0 {
						dueText += fmt.Sprintf(" (in %d days)", daysUntil)
					} else {
						dueText += fmt.Sprintf(" (%d days ago)", -daysUntil)
					}

					fmt.Fprintf(&buf, "%s %s\n", labelStyle.Render("Due:"), dueText)
				}
			}

			// Show creator
			if todo.Creator.Name != "" {
				fmt.Fprintf(&buf, "%s %s\n", labelStyle.Render("Created by:"), todo.Creator.Name)
			}

			// Show assignees
			if len(todo.Assignees) > 0 {
				var assigneeNames []string
				for _, assignee := range todo.Assignees {
					assigneeNames = append(assigneeNames, assignee.Name)
				}
				fmt.Fprintf(&buf, "%s %s\n", labelStyle.Render("Assigned to:"), strings.Join(assigneeNames, ", "))
			}

			// Show description if present
			if todo.Description != "" {
				fmt.Fprintln(&buf)
				fmt.Fprintln(&buf, labelStyle.Render("Description:"))

				// Try to render as markdown
				r, err := glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(80),
				)
				if err == nil {
					rendered, err := r.Render(todo.Description)
					if err == nil {
						fmt.Fprint(&buf, rendered)
					} else {
						fmt.Fprintln(&buf, todo.Description)
					}
				} else {
					fmt.Fprintln(&buf, todo.Description)
				}
			}

			// Show todo list ID
			if todo.TodolistID > 0 {
				fmt.Fprintln(&buf)
				fmt.Fprintf(&buf, "%s %d\n", labelStyle.Render("Todo List ID:"), todo.TodolistID)
			}

			// Show timestamps
			fmt.Fprintln(&buf)
			if created, err := time.Parse(time.RFC3339, todo.CreatedAt); err == nil {
				fmt.Fprintf(&buf, "%s %s\n", labelStyle.Render("Created:"), created.Format("January 2, 2006 at 3:04 PM"))
			}
			if updated, err := time.Parse(time.RFC3339, todo.UpdatedAt); err == nil {
				fmt.Fprintf(&buf, "%s %s\n", labelStyle.Render("Updated:"), updated.Format("January 2, 2006 at 3:04 PM"))
			}

			fmt.Fprintln(&buf)

			// Display using pager
			pagerOpts := &utils.PagerOptions{
				Pager:   cfg.Preferences.Pager,
				NoPager: noPager,
			}
			return utils.ShowInPager(buf.String(), pagerOpts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "", "Output format (json)")
	cmd.Flags().StringVar(&jsonFields, "json-fields", "", "Comma-separated list of JSON fields to output")
	cmd.Flags().BoolVarP(&webView, "web", "w", false, "Open in browser")
	cmd.Flags().BoolVar(&noPager, "no-pager", false, "Disable pager for output")

	return cmd
}
