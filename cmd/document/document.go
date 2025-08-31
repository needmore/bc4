package document

import (
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewDocumentCmd creates a new document command
func NewDocumentCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "document",
		Short:   "Work with Basecamp documents",
		Long:    `Create, edit and view documents in Basecamp projects.`,
		Aliases: []string{"documents", "doc", "docs"},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newEditCmd(f))

	return cmd
}
