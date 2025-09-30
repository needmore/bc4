package comment

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view <comment-id>",
		Short: "View a comment",
		Long:  `View the details of a specific comment.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			// Parse comment ID
			commentID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid comment ID: %s", args[0])
			}

			// Get project ID
			projectID, err := f.ProjectID()
			if err != nil {
				return err
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

			// Show in pager
			pagerOpts := &utils.PagerOptions{
				Pager:   cfg.Preferences.Pager,
				NoPager: noPager,
			}

			return utils.ShowInPager(buf.String(), pagerOpts)
		},
	}

	cmd.Flags().BoolVar(&noPager, "no-pager", false, "Don't use a pager")

	return cmd
}
