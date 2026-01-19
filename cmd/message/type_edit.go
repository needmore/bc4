package message

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newTypeEditCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var name string
	var icon string
	var color string

	cmd := &cobra.Command{
		Use:   "edit [CATEGORY_ID or URL]",
		Short: "Edit a message category",
		Long: `Edit an existing message category.

You can update the name, icon (emoji), and/or color of a category.
At least one flag must be provided to update.

Examples:
  bc4 message type edit 123 --name "Important Announcements"
  bc4 message type edit 123 --icon "📣" --color "red"
  bc4 message type edit 123 --name "Questions" --icon "❓"`,
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

			// Check if at least one flag is provided
			if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("icon") && !cmd.Flags().Changed("color") {
				return fmt.Errorf("at least one flag (--name, --icon, --color) must be provided")
			}

			// Update the category
			category, err := client.UpdateMessageCategory(
				f.Context(),
				resolvedProjectID,
				categoryID,
				name,
				icon,
				color,
			)
			if err != nil {
				return fmt.Errorf("failed to update message category: %w", err)
			}

			// Output success message
			fmt.Printf("Updated message category: %s (ID: %d)\n", category.Name, category.ID)
			if category.Icon != "" {
				fmt.Printf("Icon: %s\n", category.Icon)
			}
			if category.Color != "" {
				fmt.Printf("Color: %s\n", category.Color)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Category name")
	cmd.Flags().StringVar(&icon, "icon", "", "Category icon (emoji)")
	cmd.Flags().StringVar(&color, "color", "", "Category color")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
