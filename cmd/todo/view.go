package todo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	attachmentsCmd "github.com/needmore/bc4/cmd/attachments"
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var formatStr string
	var jsonFields string
	var webView bool
	var noPager bool
	var withComments bool

	cmd := &cobra.Command{
		Use:   "view <todo-id|url>",
		Short: "View details of a specific todo",
		Long: `View detailed information about a specific todo item including description, assignees, due date, and completion status.

You can specify the todo using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/todos/12345")`,
		Example: `bc4 todo view 12345
bc4 todo view https://3.basecamp.com/.../todos/12345
bc4 todo view 12345 --with-comments`,
		Args: cmdutil.ExactArgs(1, "todo-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply account override if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Apply project override if specified
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Parse todo ID (could be numeric ID or URL)
			todoID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid todo ID or URL: %s", args[0])
			}

			// If a URL was parsed, override account and project IDs if provided
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeTodo {
					return fmt.Errorf("URL is not for a todo: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
				}
				if parsedURL.ProjectID > 0 {
					f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
				}
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			todoOps := client.Todos()

			// Get resolved IDs
			resolvedAccountID, err := f.AccountID()
			if err != nil {
				return err
			}

			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Fetch the todo details
			todo, err := todoOps.GetTodo(f.Context(), resolvedProjectID, todoID)
			if err != nil {
				return fmt.Errorf("failed to get todo: %w", err)
			}

			// Open in browser if requested
			if webView {
				// For now, just print the URL since OpenInBrowser is not implemented
				fmt.Printf("Todo URL: https://3.basecamp.com/%s/buckets/%s/todos/%d\n", resolvedAccountID, resolvedProjectID, todoID)
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
					_ = json.Unmarshal(todoJSON, &todoMap)

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

			// Handle AI-optimized markdown output with comments
			if withComments {
				comments, err := client.ListComments(f.Context(), resolvedProjectID, todo.ID)
				if err != nil {
					return fmt.Errorf("failed to fetch comments: %w", err)
				}

				markdown, err := utils.FormatTodoAsMarkdown(todo, comments)
				if err != nil {
					return fmt.Errorf("failed to format todo as markdown: %w", err)
				}

				fmt.Print(markdown)
				return nil
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

			// Show attachments if present
			if todo.Description != "" {
				attachmentInfo := attachmentsCmd.DisplayAttachmentsWithStyle(todo.Description)
				if attachmentInfo != "" {
					fmt.Fprint(&buf, attachmentInfo)
				}
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
			cfg, _ := f.Config()
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
	cmd.Flags().BoolVar(&withComments, "with-comments", false, "Display all comments inline")

	return cmd
}
