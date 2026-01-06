package message

import (
	"fmt"
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newUnpinCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpin <message-id|url>",
		Short: "Unpin a message from the top of the message board",
		Long:  `Unpin a message so it no longer appears at the top of the message board.`,
		Example: `bc4 message unpin 12345
bc4 message unpin https://3.basecamp.com/.../messages/12345`,
		Args: cmdutil.ExactArgs(1, "message-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			var messageID int64
			var projectID string

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
				messageID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid message ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			if err := client.UnpinMessage(f.Context(), projectID, messageID); err != nil {
				return err
			}

			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("Unpinned message #%d\n", messageID)
			}

			return nil
		},
	}

	return cmd
}
