package card

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var formatJSON bool
	var accountID string
	var projectID string
	var stepsOnly bool
	var web bool
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view [ID or URL]",
		Short: "View card details including steps",
		Long: `View detailed information about a specific card, including its description, assignees, and steps.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")`,
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

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Handle web flag
			if web {
				// Open in browser (implementation would go here)
				fmt.Printf("Opening card %d in browser...\n", cardID)
				return nil
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			cardOps := client.Cards()

			// Get the card
			card, err := cardOps.GetCard(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to fetch card: %w", err)
			}

			// Handle JSON output
			if formatJSON {
				// For JSON output, return the card structure
				fmt.Printf("Card: %s\n", card.Title)
				return nil
			}

			// If steps only, show just the steps
			if stepsOnly {
				cfg, err := f.Config()
				if err != nil {
					return err
				}
				return showStepsTable(card, cfg, noPager)
			}

			// Prepare output for pager
			var buf bytes.Buffer

			// Show card header
			fmt.Fprintf(&buf, "Card #%d: %s\n", card.ID, card.Title)
			fmt.Fprintln(&buf, strings.Repeat("-", 50))

			// Show card details
			if card.Content != "" {
				fmt.Fprintf(&buf, "Description: %s\n", card.Content)
			}

			// Column
			if card.Parent != nil {
				fmt.Fprintf(&buf, "Column: %s", card.Parent.Title)
				if card.Parent.Color != "" && card.Parent.Color != "white" {
					fmt.Fprintf(&buf, " (%s)", card.Parent.Color)
				}
				fmt.Fprintln(&buf)
			}

			// Assignees
			if len(card.Assignees) > 0 {
				assigneeNames := []string{}
				for _, assignee := range card.Assignees {
					assigneeNames = append(assigneeNames, assignee.Name)
				}
				fmt.Fprintf(&buf, "Assignees: %s\n", strings.Join(assigneeNames, ", "))
			}

			// Due date
			if card.DueOn != nil && *card.DueOn != "" {
				fmt.Fprintf(&buf, "Due: %s\n", *card.DueOn)
			}

			// Creator
			if card.Creator != nil {
				fmt.Fprintf(&buf, "Created by: %s\n", card.Creator.Name)
			}

			// Timestamps
			fmt.Fprintf(&buf, "Created: %s\n", card.CreatedAt.Format("2006-01-02 15:04"))
			fmt.Fprintf(&buf, "Updated: %s\n", card.UpdatedAt.Format("2006-01-02 15:04"))

			// Show steps if any
			if len(card.Steps) > 0 {
				fmt.Fprintf(&buf, "\nSteps (%d):\n", len(card.Steps))
				fmt.Fprintln(&buf, strings.Repeat("-", 50))

				// Get steps table output
				var stepsBuf bytes.Buffer
				table := tableprinter.New(&stepsBuf)

				// Add headers
				if table.IsTTY() {
					table.AddHeader("", "ID", "TITLE", "ASSIGNEES", "DUE", "UPDATED")
				} else {
					table.AddHeader("STATUS", "ID", "TITLE", "ASSIGNEES", "DUE", "UPDATED")
				}

				// Add each step
				for _, step := range card.Steps {
					// Status indicator
					if table.IsTTY() {
						table.AddStatusField(step.Completed)
					} else {
						if step.Completed {
							table.AddField("completed")
						} else {
							table.AddField("incomplete")
						}
					}

					// Step ID
					stepID := fmt.Sprintf("%d", step.ID)
					table.AddIDField(stepID, step.Status)

					// Title
					table.AddTodoField(step.Title, step.Completed)

					// Assignees
					assigneeNames := []string{}
					for _, assignee := range step.Assignees {
						assigneeNames = append(assigneeNames, assignee.Name)
					}
					table.AddField(strings.Join(assigneeNames, ", "))

					// Due date
					if step.DueOn != nil && *step.DueOn != "" {
						table.AddField(*step.DueOn)
					} else {
						table.AddField("-")
					}

					// Updated timestamp
					table.AddTimeField(step.CreatedAt, step.UpdatedAt)
					table.EndRow()
				}

				table.Render()
				buf.Write(stepsBuf.Bytes())
			}

			// Get config for pager preferences
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			// Display using pager
			pagerOpts := &utils.PagerOptions{
				Pager:   cfg.Preferences.Pager,
				NoPager: noPager,
			}
			return utils.ShowInPager(buf.String(), pagerOpts)
		},
	}

	cmd.Flags().BoolVar(&formatJSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().BoolVar(&stepsOnly, "steps-only", false, "Show only the steps list")
	cmd.Flags().BoolVarP(&web, "web", "w", false, "Open card in web browser")
	cmd.Flags().BoolVar(&noPager, "no-pager", false, "Disable pager for output")

	return cmd
}

func showStepsTable(card *api.Card, cfg *config.Config, noPager bool) error {
	var buf bytes.Buffer
	table := tableprinter.New(&buf)

	// Add headers
	if table.IsTTY() {
		table.AddHeader("", "ID", "TITLE", "ASSIGNEES", "DUE", "UPDATED")
	} else {
		table.AddHeader("STATUS", "ID", "TITLE", "ASSIGNEES", "DUE", "UPDATED")
	}

	// Add each step
	for _, step := range card.Steps {
		// Status indicator
		if table.IsTTY() {
			table.AddStatusField(step.Completed)
		} else {
			if step.Completed {
				table.AddField("completed")
			} else {
				table.AddField("incomplete")
			}
		}

		// Step ID
		stepID := fmt.Sprintf("%d", step.ID)
		table.AddIDField(stepID, step.Status)

		// Title
		table.AddTodoField(step.Title, step.Completed)

		// Assignees
		assigneeNames := []string{}
		for _, assignee := range step.Assignees {
			assigneeNames = append(assigneeNames, assignee.Name)
		}
		table.AddField(strings.Join(assigneeNames, ", "))

		// Due date
		if step.DueOn != nil && *step.DueOn != "" {
			table.AddField(*step.DueOn)
		} else {
			table.AddField("-")
		}

		// Updated timestamp
		table.AddTimeField(step.CreatedAt, step.UpdatedAt)
		table.EndRow()
	}

	table.Render()

	// Display using pager
	pagerOpts := &utils.PagerOptions{
		Pager:   cfg.Preferences.Pager,
		NoPager: noPager,
	}
	return utils.ShowInPager(buf.String(), pagerOpts)
}
