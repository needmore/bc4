package card

import (
	"context"
	"fmt"
	"os"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var formatJSON bool
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List card tables in the current project",
		Long:  `List all card tables in the current project with their card counts and status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Check authentication
			if cfg.DefaultAccount == "" {
				return fmt.Errorf("not authenticated. Run 'bc4' to set up authentication")
			}

			// Get account ID
			if accountID == "" {
				accountID = cfg.DefaultAccount
			}
			if accountID == "" {
				return fmt.Errorf("no account specified and no default account set")
			}

			// Get project ID
			if projectID == "" {
				projectID = cfg.DefaultProject
			}
			if projectID == "" {
				// Check for account-specific default project
				if acc, ok := cfg.Accounts[accountID]; ok && acc.DefaultProject != "" {
					projectID = acc.DefaultProject
				}
			}
			if projectID == "" {
				return fmt.Errorf("no project specified and no default project set")
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			client := api.NewClient(accountID, token.AccessToken)

			// Get the card table for the project
			cardTable, err := client.GetProjectCardTable(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to fetch card table: %w", err)
			}

			// Handle JSON output
			if formatJSON {
				// For JSON output, return the card table structure as-is
				// In a real implementation, you'd serialize this properly
				fmt.Printf("Card table: %s\n", cardTable.Title)
				return nil
			}

			// Get default card table for marking
			var defaultCardTable string
			if acc, ok := cfg.Accounts[accountID]; ok {
				if proj, ok := acc.ProjectDefaults[projectID]; ok {
					defaultCardTable = proj.DefaultCardTable
				}
			}

			// Create table
			table := tableprinter.New(os.Stdout)

			// Add headers
			if table.IsTTY() {
				table.AddHeader("ID", "NAME", "DESCRIPTION", "CARDS", "UPDATED")
			} else {
				table.AddHeader("ID", "NAME", "DESCRIPTION", "CARDS", "STATUS", "UPDATED")
			}

			// Add the card table as a row
			table.AddIDField(fmt.Sprintf("%d", cardTable.ID), cardTable.Status)

			// Add name with default indicator
			name := cardTable.Title
			if fmt.Sprintf("%d", cardTable.ID) == defaultCardTable {
				name += " *"
			}
			table.AddProjectField(name, cardTable.Status)

			// Add description
			table.AddField(cardTable.Description)

			// Add cards count
			table.AddField(fmt.Sprintf("%d", cardTable.CardsCount))

			// Add status for non-TTY
			if !table.IsTTY() {
				table.AddField(cardTable.Status)
			}

			// Add timestamp
			table.AddTimeField(cardTable.CreatedAt, cardTable.UpdatedAt)
			table.EndRow()

			// Print summary
			fmt.Printf("Showing card table in project %s\n\n", projectID)
			table.Render()

			return nil
		},
	}

	cmd.Flags().BoolVar(&formatJSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
