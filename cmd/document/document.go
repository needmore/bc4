package document

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewDocumentCmd creates a new document command
func NewDocumentCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "document",
		Short:   "Work with Basecamp documents",
		Aliases: []string{"documents", "doc", "docs"},
		Long: `Create, edit, and view documents in Basecamp projects.

Documents are great for sharing specifications, guidelines, meeting notes,
or any other long-form content with your team. They support rich text
formatting and allow for collaborative editing and commenting.`,
		Example: `  bc4 document list                   # List all documents
  bc4 document create "Spec"          # Create a new document
  bc4 document view 123               # View document content
  bc4 document edit 123               # Edit a document`,
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newDownloadAttachmentsCmd(f))

	return cmd
}
