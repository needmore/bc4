package comment

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	attachmentsCmd "github.com/needmore/bc4/cmd/attachments"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view <comment-id|url>",
		Short: "View a comment",
		Long:  `View the details of a specific comment.`,
		Args:  cobra.ExactArgs(1),
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

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			var commentID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeComment {
					return fmt.Errorf("URL is not a comment URL: %s", args[0])
				}
				commentID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				commentID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid comment ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Get the comment
			comment, err := client.GetComment(f.Context(), projectID, commentID)
			if err != nil {
				return err
			}

			// Build formatted output
			var buf bytes.Buffer

			// Title
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
			fmt.Fprintf(&buf, "\n%s\n", titleStyle.Render(fmt.Sprintf("Comment #%d", comment.ID)))

			// Metadata
			metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
			fmt.Fprintf(&buf, "%s\n", metaStyle.Render(fmt.Sprintf("By %s • %s",
				comment.Creator.Name,
				comment.CreatedAt.Format("Jan 2, 2006"))))

			if comment.Parent.Title != "" {
				fmt.Fprintf(&buf, "%s\n", metaStyle.Render(fmt.Sprintf("On: %s (%s)", comment.Parent.Title, comment.Parent.Type)))
			}

			fmt.Fprintln(&buf)

			// Render content with glamour
			r, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(80),
			)
			if err != nil {
				return fmt.Errorf("failed to create renderer: %w", err)
			}

			rendered, err := r.Render(comment.Content)
			if err != nil {
				return fmt.Errorf("failed to render content: %w", err)
			}

			fmt.Fprint(&buf, rendered)

			// Show attachments if present
			if comment.Content != "" {
				attachmentInfo := attachmentsCmd.DisplayAttachmentsWithStyle(comment.Content)
				if attachmentInfo != "" {
					fmt.Fprint(&buf, attachmentInfo)
				}
			}

			// Show in pager
			pagerOpts := &utils.PagerOptions{
				Pager:   cfg.Preferences.Pager,
				NoPager: noPager,
			}

			return utils.ShowInPager(buf.String(), pagerOpts)
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().BoolVar(&noPager, "no-pager", false, "Don't use a pager")

	return cmd
}
