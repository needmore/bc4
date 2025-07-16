package card

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

// newStepListCmd creates the step list command
func newStepListCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "list [CARD_ID or URL]",
		Short: "List all steps in a card",
		Long: `List all steps (subtasks) in a card, showing their status and assignees.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Examples:
  bc4 card step list 123
  bc4 card step list 123 --completed
  bc4 card step list 123 --format json
  bc4 card step list https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// If a URL was parsed, override account and project IDs if provided
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeCard {
					return fmt.Errorf("URL is not for a card: %s", args[0])
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

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get the card with its steps
			card, err := client.Cards().GetCard(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to get card: %w", err)
			}

			// Apply filters
			showCompleted, _ := cmd.Flags().GetBool("completed")
			showPending, _ := cmd.Flags().GetBool("pending")
			assigneeFilter, _ := cmd.Flags().GetString("assignee")

			// Filter steps based on flags
			filteredSteps := card.Steps
			if showCompleted || showPending {
				var filtered []api.Step
				for _, step := range card.Steps {
					if (showCompleted && step.Status == "completed") ||
						(showPending && step.Status != "completed") {
						filtered = append(filtered, step)
					}
				}
				filteredSteps = filtered
			}

			// Filter by assignee if specified
			if assigneeFilter != "" {
				var filtered []api.Step
				for _, step := range filteredSteps {
					for _, assignee := range step.Assignees {
						if strings.Contains(strings.ToLower(assignee.Name), strings.ToLower(assigneeFilter)) ||
							strings.Contains(strings.ToLower(assignee.EmailAddress), strings.ToLower(assigneeFilter)) {
							filtered = append(filtered, step)
							break
						}
					}
				}
				filteredSteps = filtered
			}

			// Handle different output formats
			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				output, err := json.MarshalIndent(filteredSteps, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(string(output))

			case "tsv":
				// Output tab-separated values
				fmt.Println("ID\tTITLE\tSTATUS\tASSIGNEES\tDUE_ON")
				for _, step := range filteredSteps {
					assigneeNames := []string{}
					for _, a := range step.Assignees {
						assigneeNames = append(assigneeNames, a.Name)
					}
					dueOn := ""
					if step.DueOn != nil {
						dueOn = *step.DueOn
					}
					fmt.Printf("%d\t%s\t%s\t%s\t%s\n",
						step.ID,
						step.Title,
						step.Status,
						strings.Join(assigneeNames, ", "),
						dueOn)
				}

			default: // table format
				table := tableprinter.New(os.Stdout)

				// Add headers
				if table.IsTTY() {
					table.AddHeader("", "ID", "TITLE", "ASSIGNEES", "DUE")
				} else {
					table.AddHeader("ID", "TITLE", "STATUS", "ASSIGNEES", "DUE_ON")
				}

				// Add rows
				for _, step := range filteredSteps {
					if table.IsTTY() {
						// Visual status indicator
						status := "○"
						if step.Status == "completed" {
							status = "✓"
						}
						table.AddField(status)
						table.AddIDField(fmt.Sprintf("%d", step.ID), step.Status)
					} else {
						table.AddField(fmt.Sprintf("#%d", step.ID))
					}

					table.AddField(step.Title)

					if !table.IsTTY() {
						table.AddField(step.Status)
					}

					// Assignees
					assigneeNames := []string{}
					for _, a := range step.Assignees {
						assigneeNames = append(assigneeNames, a.Name)
					}
					table.AddField(strings.Join(assigneeNames, ", "))

					// Due date
					if step.DueOn != nil && *step.DueOn != "" {
						table.AddField(*step.DueOn)
					} else {
						table.AddField("")
					}

					table.EndRow()
				}

				_ = table.Render()
			}

			return nil
		},
	}

	// TODO: Add flags for filtering and formatting
	cmd.Flags().Bool("completed", false, "Show only completed steps")
	cmd.Flags().Bool("pending", false, "Show only pending steps")
	cmd.Flags().String("assignee", "", "Filter by assignee")
	cmd.Flags().String("format", "table", "Output format: table, json, csv")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
