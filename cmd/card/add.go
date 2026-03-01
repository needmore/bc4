package card

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/attachments"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/mentions"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

func newAddCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var tableID string
	var columnName string
	var assignees []string
	var steps []string
	var dueOn string
	var description string
	var attach []string

	cmd := &cobra.Command{
		Use:   "add \"Title\"",
		Short: "Quick card creation",
		Long: `Create a new card in the default card table's first column.

Use flags to specify table, column, assignees, and initial steps.

Use --attach to add images or files to the card content. Multiple files
can be attached by using the flag multiple times.`,
		Example: `  # Create a simple card
  bc4 card add "New feature"

  # Create a card with description
  bc4 card add "Bug fix" --description "Fix login issue"

  # Create a card with an image attachment
  bc4 card add "Design review" --attach ./mockup.png

  # Create a card with multiple attachments
  bc4 card add "Asset update" --attach ./logo.png --attach ./banner.jpg`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]

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
			cardOps := client.Cards()
			stepOps := client.Steps()

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get config for default lookups
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			// Get resolved account ID for defaults
			resolvedAccountID, err := f.AccountID()
			if err != nil {
				return err
			}

			// Get card table ID
			var cardTableID int64
			if tableID != "" {
				// Check if it's a URL
				if parser.IsBasecampURL(tableID) {
					parsed, err := parser.ParseBasecampURL(tableID)
					if err != nil {
						return fmt.Errorf("invalid Basecamp URL: %w", err)
					}
					if parsed.ResourceType != parser.ResourceTypeCardTable {
						return fmt.Errorf("URL is not a card table URL: %s", tableID)
					}
					cardTableID = parsed.ResourceID
				} else {
					// Parse specified table ID
					if id, err := strconv.ParseInt(tableID, 10, 64); err == nil {
						cardTableID = id
					} else {
						// Search by name not implemented yet
						return fmt.Errorf("searching card tables by name not yet implemented")
					}
				}
			} else {
				// Use default card table
				if acc, ok := cfg.Accounts[resolvedAccountID]; ok {
					if proj, ok := acc.ProjectDefaults[resolvedProjectID]; ok && proj.DefaultCardTable != "" {
						if id, err := strconv.ParseInt(proj.DefaultCardTable, 10, 64); err == nil {
							cardTableID = id
						}
					}
				}
				if cardTableID == 0 {
					// No default set, get the project's card table
					cardTable, err := cardOps.GetProjectCardTable(f.Context(), resolvedProjectID)
					if err != nil {
						return fmt.Errorf("failed to fetch card table: %w", err)
					}
					cardTableID = cardTable.ID
				}
			}

			// Get the card table to find columns
			cardTable, err := cardOps.GetCardTable(f.Context(), resolvedProjectID, cardTableID)
			if err != nil {
				return fmt.Errorf("failed to fetch card table: %w", err)
			}

			// Find the target column
			var targetColumn *api.Column
			if columnName != "" {
				// Find column by name
				for i := range cardTable.Lists {
					if strings.Contains(strings.ToLower(cardTable.Lists[i].Title), strings.ToLower(columnName)) {
						targetColumn = &cardTable.Lists[i]
						break
					}
				}
				if targetColumn == nil {
					return fmt.Errorf("column '%s' not found", columnName)
				}
			} else {
				// Use first non-triage column (usually the second column)
				for i := range cardTable.Lists {
					if cardTable.Lists[i].Type != "Kanban::Triage" {
						targetColumn = &cardTable.Lists[i]
						break
					}
				}
				if targetColumn == nil && len(cardTable.Lists) > 0 {
					// Fallback to first column
					targetColumn = &cardTable.Lists[0]
				}
			}

			if targetColumn == nil {
				return fmt.Errorf("no suitable column found in card table")
			}

			// Convert description to rich text if provided
			var richContent string
			if description != "" {
				converter := markdown.NewConverter()
				rc, err := converter.MarkdownToRichText(description)
				if err != nil {
					return fmt.Errorf("failed to convert description: %w", err)
				}
				richContent = rc
			}

			// Replace inline @Name mentions with bc-attachment tags
			if richContent != "" {
				richContent, err = mentions.Resolve(f.Context(), richContent, client.Client, resolvedProjectID)
				if err != nil {
					return fmt.Errorf("failed to resolve mentions: %w", err)
				}
			}

			// Handle attachments
			if len(attach) > 0 {
				for _, attachPath := range attach {
					fileData, err := os.ReadFile(attachPath)
					if err != nil {
						return fmt.Errorf("failed to read attachment %s: %w", attachPath, err)
					}
					filename := filepath.Base(attachPath)
					upload, err := client.UploadAttachment(filename, fileData, "")
					if err != nil {
						return fmt.Errorf("failed to upload attachment %s: %w", filename, err)
					}
					tag := attachments.BuildTag(upload.AttachableSGID)
					richContent += tag
				}
			}

			// Create the card
			req := api.CardCreateRequest{
				Title:   title,
				Content: richContent,
			}
			if dueOn != "" {
				req.DueOn = &dueOn
			}

			card, err := cardOps.CreateCard(f.Context(), resolvedProjectID, targetColumn.ID, req)
			if err != nil {
				return fmt.Errorf("failed to create card: %w", err)
			}

			fmt.Printf("Created card #%d: %s in column '%s'\n", card.ID, card.Title, targetColumn.Title)

			// Handle assignees - resolve user identifiers
			if len(assignees) > 0 {
				// Create user resolver
				userResolver := utils.NewUserResolver(client.Client, resolvedProjectID)

				// Resolve user identifiers to person IDs
				assigneeIDs, err := userResolver.ResolveUsers(f.Context(), assignees)
				if err != nil {
					fmt.Printf("Warning: failed to resolve assignees: %v\n", err)
				} else if len(assigneeIDs) > 0 {
					updateReq := api.CardUpdateRequest{
						AssigneeIDs: assigneeIDs,
					}
					_, err := cardOps.UpdateCard(f.Context(), resolvedProjectID, card.ID, updateReq)
					if err != nil {
						fmt.Printf("Warning: failed to assign users: %v\n", err)
					} else {
						fmt.Printf("Assigned %d user(s) to the card\n", len(assigneeIDs))
					}
				}
			}

			// Add steps if provided
			if len(steps) > 0 {
				fmt.Printf("Adding %d steps...\n", len(steps))
				for _, stepTitle := range steps {
					stepReq := api.StepCreateRequest{
						Title: stepTitle,
					}
					_, err := stepOps.CreateStep(f.Context(), resolvedProjectID, card.ID, stepReq)
					if err != nil {
						fmt.Printf("Warning: failed to add step '%s': %v\n", stepTitle, err)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&tableID, "table", "", "Specify card table ID or URL")
	cmd.Flags().StringVar(&columnName, "column", "", "Target column name")
	cmd.Flags().StringSliceVar(&assignees, "assign", []string{}, "Add assignees by email or @mention (comma-separated)")
	cmd.Flags().StringSliceVar(&steps, "step", []string{}, "Add steps (can be used multiple times)")
	cmd.Flags().StringVar(&dueOn, "due", "", "Set due date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Card description")
	cmd.Flags().StringSliceVar(&attach, "attach", nil, "Attach file(s) to the card (can be used multiple times)")

	return cmd
}
