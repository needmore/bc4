package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/needmore/bc4/cmd/account"
	"github.com/needmore/bc4/cmd/auth"
	"github.com/needmore/bc4/cmd/campfire"
	"github.com/needmore/bc4/cmd/card"
	"github.com/needmore/bc4/cmd/comment"
	"github.com/needmore/bc4/cmd/document"
	"github.com/needmore/bc4/cmd/message"
	"github.com/needmore/bc4/cmd/project"
	"github.com/needmore/bc4/cmd/todo"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/errors"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/tui"
	"github.com/needmore/bc4/internal/version"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "bc4",
	Short:   "A CLI tool for interacting with Basecamp 4",
	Version: version.Get().Version,
	Long: `bc4 is a command-line interface for Basecamp 4 that allows you to:
• List and manage projects and accounts
• Create and manage todos and todo lists
• Post and edit messages and documents
• Add and manage comments on any resource
• Work with campfires (team chat)
• Manage card tables (kanban boards)
• And much more!

Get started by running 'bc4' to launch the setup wizard.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if this is the first run
		if config.IsFirstRun() {
			// Run first-run wizard immediately with clean screen
			p := tea.NewProgram(
				tui.NewFirstRunModel(),
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}

			fmt.Println("\nSetup complete! You can now use bc4.")
			fmt.Println("Try 'bc4 auth status' to see your account information.")
			return nil
		}

		// Show help if no subcommand
		return cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// Don't format cobra's built-in errors (help, version, etc.)
		// These are displayed properly by cobra itself
		if err.Error() == "help requested" {
			os.Exit(0)
		}

		// Format the error for user display
		fmt.Fprintln(os.Stderr, errors.FormatError(err))
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/bc4/config.json)")
	rootCmd.PersistentFlags().StringP("account", "a", "", "Override default account ID")
	rootCmd.PersistentFlags().StringP("project", "p", "", "Override default project ID")
	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")

	// Bind flags to viper
	_ = viper.BindPFlag("account", rootCmd.PersistentFlags().Lookup("account"))
	_ = viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))

	// Create factory
	f := factory.New()

	// Add commands with factory
	rootCmd.AddCommand(auth.NewAuthCmd(f))
	rootCmd.AddCommand(account.NewAccountCmd(f))
	rootCmd.AddCommand(project.NewProjectCmd(f))
	rootCmd.AddCommand(todo.NewTodoCmd(f))
	rootCmd.AddCommand(message.NewMessageCmd(f))
	rootCmd.AddCommand(document.NewDocumentCmd(f))
	rootCmd.AddCommand(campfire.NewCampfireCmd(f))
	rootCmd.AddCommand(card.NewCardCmd(f))
	rootCmd.AddCommand(comment.NewCommentCmd(f))

	// Add version command (doesn't need factory)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	// Set config file if specified
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Use default config location
		configDir, err := os.UserConfigDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(configDir + "/bc4")
		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}

	// Environment variables
	viper.SetEnvPrefix("BC4")
	viper.AutomaticEnv()

	// Read config
	_ = viper.ReadInConfig()
}
