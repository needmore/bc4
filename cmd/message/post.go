package message

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/mentions"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newPostCmd(f *factory.Factory) *cobra.Command {
	var (
		title      string
		content    string
		categoryID int64
	)

	cmd := &cobra.Command{
		Use:   "post [project]",
		Short: "Post a new message",
		Long: `Post a new message to a project's message board.

You can provide message content in several ways:
  - Interactively (default)
  - Via --content flag
  - Via stdin: echo "content" | bc4 message post [project] --title "Title"
  - From file: cat message.md | bc4 message post [project] --title "Title"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply project override if specified
			if len(args) > 0 {
				f = f.WithProject(args[0])
			}

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
			board, err := client.GetMessageBoard(f.Context(), projectID)
			if err != nil {
				return err
			}

			// Check if stdin has data
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// Data is being piped in
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				content = strings.TrimSpace(string(data))

				// Title is required when using stdin
				if title == "" {
					return fmt.Errorf("--title is required when piping content via stdin")
				}
			} else if content == "" {
				// No stdin and no content flag, use interactive mode
				if title == "" {
					if err := huh.NewInput().
						Title("Message subject").
						Placeholder("What's this message about?").
						Value(&title).
						Run(); err != nil {
						return err
					}
				}

				if err := huh.NewText().
					Title("Message content").
					Placeholder("Write your message in Markdown...").
					Lines(10).
					Value(&content).
					Run(); err != nil {
					return err
				}
			}

			// Validate required fields
			if title == "" {
				return fmt.Errorf("message subject is required")
			}
			if content == "" {
				return fmt.Errorf("message content is required")
			}

			// Convert markdown to rich text
			converter := markdown.NewConverter()
			richContent, err := converter.MarkdownToRichText(content)
			if err != nil {
				return fmt.Errorf("failed to convert markdown: %w", err)
			}

			// Replace inline @Name mentions with bc-attachment tags
			richContent, err = mentions.Resolve(f.Context(), richContent, client.Client, projectID)
			if err != nil {
				return fmt.Errorf("failed to resolve mentions: %w", err)
			}

			// Create the message
			req := api.MessageCreateRequest{
				Subject: title,
				Content: richContent,
				Status:  "active",
			}

			if categoryID > 0 {
				req.CategoryID = &categoryID
			}

			message, err := client.CreateMessage(f.Context(), projectID, board.ID, req)
			if err != nil {
				return err
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Created message #%d: %s\n", message.ID, message.Subject)
			} else {
				fmt.Println(message.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "Message subject")
	cmd.Flags().StringVarP(&content, "content", "c", "", "Message content (markdown supported)")
	cmd.Flags().Int64Var(&categoryID, "category-id", 0, "Category ID")

	return cmd
}
