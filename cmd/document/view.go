package document

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/glamour"

	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var withComments bool
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view <document-id|url>",
		Short: "View a document",
		Long:  `View a specific document by ID or URL.`,
		Example: `bc4 document view 12345
bc4 document view https://3.basecamp.com/.../documents/12345`,
		Args: cmdutil.ExactArgs(1, "document-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			var documentID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeDocument {
					return fmt.Errorf("URL is not a document URL: %s", args[0])
				}
				documentID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				documentID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid document ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Get the document
			document, err := client.GetDocument(f.Context(), projectID, documentID)
			if err != nil {
				return err
			}

			// Handle output with comments
			if withComments {
				comments, err := client.ListComments(f.Context(), projectID, document.ID)
				if err != nil {
					return fmt.Errorf("failed to fetch comments: %w", err)
				}

				mdContent, err := utils.FormatDocumentAsMarkdown(document, comments)
				if err != nil {
					return fmt.Errorf("failed to format document as markdown: %w", err)
				}

				// If piped, output raw markdown for scripting/AI
				if !ui.IsTerminal(os.Stdout) {
					fmt.Print(mdContent)
					return nil
				}

				// Render with glamour for terminal display
				r, err := glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(80),
				)
				if err != nil {
					return fmt.Errorf("failed to create renderer: %w", err)
				}

				rendered, err := r.Render(mdContent)
				if err != nil {
					return fmt.Errorf("failed to render content: %w", err)
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
				return utils.ShowInPager(rendered, pagerOpts)
			}

			// Output format
			if viper.GetBool("json") {
				return json.NewEncoder(os.Stdout).Encode(document)
			}

			// Terminal output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("📄 %s (#%d)\n", document.Title, document.ID)
				fmt.Printf("Created: %s by %s\n", document.CreatedAt.Format("2006-01-02 15:04"), document.Creator.Name)
				if document.UpdatedAt.After(document.CreatedAt) {
					fmt.Printf("Updated: %s\n", document.UpdatedAt.Format("2006-01-02 15:04"))
				}
				if document.CommentsCount > 0 {
					fmt.Printf("Comments: %d\n", document.CommentsCount)
				}
				fmt.Printf("Status: %s\n", document.Status)
				fmt.Println()
				fmt.Println("Content:")
				fmt.Println("─────────")

				// Try to convert rich text back to markdown for better display
				converter := markdown.NewConverter()
				markdownContent, err := converter.RichTextToMarkdown(document.Content)
				if err != nil {
					// If conversion fails, show raw content
					fmt.Println(document.Content)
				} else {
					fmt.Println(markdownContent)
				}
			} else {
				// Convert to markdown for non-terminal output
				converter := markdown.NewConverter()
				markdownContent, err := converter.RichTextToMarkdown(document.Content)
				if err != nil {
					fmt.Println(document.Content)
				} else {
					fmt.Println(markdownContent)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&withComments, "with-comments", false, "Display all comments inline")
	cmd.Flags().BoolVar(&noPager, "no-pager", false, "Don't use a pager")

	return cmd
}
