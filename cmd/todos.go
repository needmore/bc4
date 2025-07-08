package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var todosCmd = &cobra.Command{
	Use:   "todos",
	Short: "Manage Basecamp todos",
	Long:  `List, create, and manage todos in Basecamp projects.`,
}

var todosListCmd = &cobra.Command{
	Use:   "list",
	Short: "List todos in a project",
	Long:  `List all todos in a specified project.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectID, _ := cmd.Flags().GetString("project")
		fmt.Printf("Listing todos for project %s... (not implemented yet)\n", projectID)
	},
}

var todosCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new todo",
	Long:  `Create a new todo in a specified project and todo list.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID, _ := cmd.Flags().GetString("project")
		listID, _ := cmd.Flags().GetString("list")
		fmt.Printf("Creating todo '%s' in project %s, list %s... (not implemented yet)\n", args[0], projectID, listID)
	},
}

func init() {
	rootCmd.AddCommand(todosCmd)
	todosCmd.AddCommand(todosListCmd)
	todosCmd.AddCommand(todosCreateCmd)
	
	todosCmd.PersistentFlags().String("project", "", "Project ID (required)")
	todosCmd.MarkPersistentFlagRequired("project")
	
	todosCreateCmd.Flags().String("list", "", "Todo list ID (required)")
	todosCreateCmd.MarkFlagRequired("list")
}