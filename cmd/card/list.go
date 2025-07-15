package card

import (
	"fmt"
	"os"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var formatJSON bool
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List card tables in the current project",
		Long:  `List all card tables in the current project with their card counts and status.`,
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
			cardOps := client.Cards()

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get the card table for the project
			cardTable, err := cardOps.GetProjectCardTable(f.Context(), resolvedProjectID)
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

			// Get config for default card table lookup
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			// Get resolved account ID for default lookup
			resolvedAccountID, err := f.AccountID()
			if err != nil {
				return err
			}

			// Get default card table for marking
			var defaultCardTable string
			if acc, ok := cfg.Accounts[resolvedAccountID]; ok {
				if proj, ok := acc.ProjectDefaults[resolvedProjectID]; ok {
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
			fmt.Printf("Showing card table in project %s\n\n", resolvedProjectID)
			table.Render()

			return nil
		},
	}

	cmd.Flags().BoolVar(&formatJSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
