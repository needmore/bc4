package document

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

func newCreateCmd(f *factory.Factory) *cobra.Command {
	var (
		title   string
		content string
		draft   bool
	)

	cmd := &cobra.Command{
		Use:   "create [project]",
		Short: "Create a new document",
		Long: `Create a new document in a project's document vault.

You can provide document content in several ways:
  - Interactively (default)
  - Via --content flag
  - Via stdin: echo "content" | bc4 document create [project] --title "Title"
  - From file: cat document.md | bc4 document create [project] --title "Title"`,
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

			// Get the vault for the project
			vault, err := client.GetVault(f.Context(), projectID)
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
						Title("Document title").
						Placeholder("What's this document about?").
						Value(&title).
						Run(); err != nil {
						return err
					}
				}

				if err := huh.NewText().
					Title("Document content").
					Placeholder("Write your document in Markdown...").
					Lines(15).
					Value(&content).
					Run(); err != nil {
					return err
				}
			}

			// Validate required fields
			if title == "" {
				return fmt.Errorf("document title is required")
			}
			if content == "" {
				return fmt.Errorf("document content is required")
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

			// Create the document
			req := api.DocumentCreateRequest{
				Title:   title,
				Content: richContent,
				Status:  "active",
			}

			if draft {
				req.Status = "draft"
			}

			document, err := client.CreateDocument(f.Context(), projectID, vault.ID, req)
			if err != nil {
				return err
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Created document #%d: %s\n", document.ID, document.Title)
			} else {
				fmt.Println(document.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "Document title")
	cmd.Flags().StringVarP(&content, "content", "c", "", "Document content (markdown supported)")
	cmd.Flags().BoolVarP(&draft, "draft", "d", false, "Create as draft")

	return cmd
}
