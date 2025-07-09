package project

import (
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the project command
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Basecamp projects",
		Long:  `Work with Basecamp projects - list, view, search, and manage projects.`,
		Aliases: []string{"p"},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSelectCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newSearchCmd())

	return cmd
}