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

func newPinCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pin <message-id|url>",
		Short: "Pin a message to the top of the message board",
		Long:  `Pin a message so it appears at the top of the message board for easy access.`,
		Example: `bc4 message pin 12345
bc4 message pin https://3.basecamp.com/.../messages/12345`,
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

			if err := client.PinMessage(f.Context(), projectID, messageID); err != nil {
				return err
			}

			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("Pinned message #%d\n", messageID)
			}

			return nil
		},
	}

	return cmd
}
