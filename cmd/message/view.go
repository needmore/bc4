package message

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view [message-id|url]",
		Short: "View a message",
		Long:  `View the details of a specific message.`,
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

			var messageID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeMessage {
					return fmt.Errorf("URL is not a message URL: %s", args[0])
				}
				messageID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				messageID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid message ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Get the message
			message, err := client.GetMessage(f.Context(), projectID, messageID)
			if err != nil {
				return err
			}

			// Build formatted output
			var buf bytes.Buffer

			// Title
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
			fmt.Fprintf(&buf, "\n%s\n", titleStyle.Render(message.Subject))

			// Metadata
			metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
			fmt.Fprintf(&buf, "%s\n", metaStyle.Render(fmt.Sprintf("By %s • %s",
				message.Creator.Name,
				message.CreatedAt.Format("Jan 2, 2006"))))

			if message.Category != nil {
				fmt.Fprintf(&buf, "%s\n", metaStyle.Render(fmt.Sprintf("Category: %s", message.Category.Name)))
			}

			if message.CommentsCount > 0 {
				fmt.Fprintf(&buf, "%s\n", metaStyle.Render(fmt.Sprintf("%d comments", message.CommentsCount)))
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

			rendered, err := r.Render(message.Content)
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
