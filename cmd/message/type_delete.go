package message

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newTypeDeleteCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "delete [CATEGORY_ID or URL]",
		Short: "Delete a message category",
		Long: `Delete a message category.

Note: Deleting a category will NOT delete messages in that category.
Messages will simply have no category assigned.

Examples:
  bc4 message type delete 123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse category ID
			categoryID, _, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid category ID or URL: %s", args[0])
			}

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
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Delete the category
			err = client.DeleteMessageCategory(
				f.Context(),
				resolvedProjectID,
				categoryID,
			)
			if err != nil {
				return fmt.Errorf("failed to delete message category: %w", err)
			}

			// Output success message
			fmt.Printf("Deleted message category (ID: %d)\n", categoryID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
