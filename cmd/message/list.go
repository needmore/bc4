package message

import (
	"fmt"
	"os"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var (
		category string
		limit    int
	)

	cmd := &cobra.Command{
		Use:     "list [project]",
		Short:   "List messages in a project",
		Long:    `List all messages on a project's message board.`,
		Aliases: []string{"ls"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get resolved project ID
			projectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get the message board for the project
			board, err := client.GetMessageBoard(cmd.Context(), projectID)
			if err != nil {
				return err
			}

			// Get all messages
			messages, err := client.ListMessages(cmd.Context(), projectID, board.ID)
			if err != nil {
				return err
			}

			// Filter by category if specified
			if category != "" {
				var filtered []api.Message
				for _, msg := range messages {
					if msg.Category != nil && msg.Category.Name == category {
						filtered = append(filtered, msg)
					}
				}
				messages = filtered
			}

			// Apply limit if specified
			if limit > 0 && len(messages) > limit {
				messages = messages[:limit]
			}

			// Display messages
			if len(messages) == 0 {
				fmt.Println("No messages found")
				return nil
			}

			// Create table
			table := tableprinter.New(os.Stdout)
			cs := table.GetColorScheme()

			// Add headers dynamically based on TTY mode
			if table.IsTTY() {
				table.AddHeader("ID", "SUBJECT", "FROM", "REPLIES", "UPDATED")
			} else {
				table.AddHeader("ID", "SUBJECT", "FROM", "CATEGORY", "REPLIES", "STATUS", "UPDATED")
			}

			// Add rows
			for _, msg := range messages {
				table.AddField(fmt.Sprintf("%d", msg.ID))
				table.AddField(msg.Subject)
				table.AddField(msg.Creator.Name, cs.Muted)

				if !table.IsTTY() {
					if msg.Category != nil {
						table.AddField(msg.Category.Name, cs.Muted)
					} else {
						table.AddField("", cs.Muted)
					}
				}

				table.AddField(fmt.Sprintf("%d", msg.CommentsCount), cs.Muted)

				if !table.IsTTY() {
					table.AddField(msg.Status, cs.Muted)
				}

				now := time.Now()
				table.AddTimeField(now, msg.UpdatedAt)
				table.EndRow()
			}

			// Render
			if err := table.Render(); err != nil {
				return fmt.Errorf("failed to render table: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category")
	cmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of messages shown")

	return cmd
}
