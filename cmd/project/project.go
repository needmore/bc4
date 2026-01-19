package project

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the project command
func NewProjectCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Manage Basecamp projects",
		Long:    `Work with Basecamp projects - list, view, search, and manage projects.`,
		Aliases: []string{"p"},
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newSelectCmd(f))
	cmd.AddCommand(newSetCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newSearchCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newDeleteCmd(f))
	cmd.AddCommand(newArchiveCmd(f))
	cmd.AddCommand(newUnarchiveCmd(f))
	cmd.AddCommand(newCopyCmd(f))

	return cmd
}
