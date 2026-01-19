package message

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newTypeCreateCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var icon string
	var color string

	cmd := &cobra.Command{
		Use:   "create [MESSAGE_BOARD_ID or URL] NAME",
		Short: "Create a new message category",
		Long: `Create a new message category for organizing messages.

You can specify the message board using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/message_boards/12345")
- Omit the ID to use the current project's message board

Categories can have optional icons (emoji) and colors for visual organization.

Examples:
  bc4 message type create "Announcements"
  bc4 message type create "Updates" --icon "📢" --color "blue"
  bc4 message type create 123 "Questions" --icon "❓"
  bc4 message type create https://3.basecamp.com/1234567/buckets/89012345/message_boards/12345 "Ideas"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var messageBoardID int64
			var categoryName string
			var err error

			// Parse arguments - could be (NAME) or (BOARD_ID, NAME)
			if len(args) == 1 {
				// Only name provided, use current project's message board
				categoryName = args[0]
			} else {
				// Both board ID and name provided
				var parsedBoardID int64
				parsedBoardID, _, err = parser.ParseArgument(args[0])
				if err != nil {
					return fmt.Errorf("invalid message board ID or URL: %s", args[0])
				}
				messageBoardID = parsedBoardID
				categoryName = args[1]
				// Note: URL parsing for message boards not currently supported in parser
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

			// Get message board if not provided
			if messageBoardID == 0 {
				board, err := client.GetMessageBoard(f.Context(), resolvedProjectID)
				if err != nil {
					return fmt.Errorf("failed to get message board: %w", err)
				}
				messageBoardID = board.ID
			}

			// Create the category
			category, err := client.CreateMessageCategory(
				f.Context(),
				resolvedProjectID,
				messageBoardID,
				categoryName,
				icon,
				color,
			)
			if err != nil {
				return fmt.Errorf("failed to create message category: %w", err)
			}

			// Output success message
			fmt.Printf("Created message category: %s (ID: %d)\n", category.Name, category.ID)
			if category.Icon != "" {
				fmt.Printf("Icon: %s\n", category.Icon)
			}
			if category.Color != "" {
				fmt.Printf("Color: %s\n", category.Color)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&icon, "icon", "", "Category icon (emoji)")
	cmd.Flags().StringVar(&color, "color", "", "Category color")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
