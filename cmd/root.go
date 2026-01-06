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
	"github.com/needmore/bc4/cmd/profile"
	"github.com/needmore/bc4/cmd/project"
	"github.com/needmore/bc4/cmd/todo"
	"github.com/needmore/bc4/internal/cmdutil"
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

Quick Start:
  bc4                        Run first-time setup wizard
  bc4 auth status            Check if authenticated
  bc4 project list           See your projects
  bc4 project select         Pick a default project
  bc4 todo lists             View todo lists
  bc4 todo list "Tasks"      See todos in a list
  bc4 todo add "New task"    Create a todo

Common Workflows:
  bc4 campfire post "Update" Quick team update
  bc4 message post           Announce to project
  bc4 card table "Bugs"      View kanban board

Shell Completions:
  bc4 completion bash        Bash
  bc4 completion zsh         Zsh
  bc4 completion fish        Fish
  bc4 completion powershell  PowerShell

See 'bc4 completion --help' for installation instructions.`,
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
			os.Exit(cmdutil.ExitSuccess)
		}

		// Unwrap silent errors to check the underlying type for exit codes
		unwrappedErr := cmdutil.UnwrapSilent(err)

		// Determine appropriate exit code based on error type
		exitCode := cmdutil.ExitError

		switch {
		case cmdutil.IsUsageError(unwrappedErr):
			exitCode = cmdutil.ExitUsageError
		case errors.IsAuthenticationError(unwrappedErr), errors.IsConfigurationError(unwrappedErr):
			exitCode = cmdutil.ExitAuthError
		case errors.IsNotFoundError(unwrappedErr):
			exitCode = cmdutil.ExitNotFound
		}

		// Only format and display error if it's not a silent error
		// (silent errors have already been displayed by the command)
		if !cmdutil.IsSilentError(err) {
			fmt.Fprintln(os.Stderr, errors.FormatError(err))
		}

		os.Exit(exitCode)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Disable Cobra's automatic usage and error printing on error
	// We handle this ourselves in Execute() for better control
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	// Enable command suggestions on typos
	cmdutil.EnableSuggestions(rootCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/bc4/config.json)")
	rootCmd.PersistentFlags().StringP("account", "a", "", "Override default account ID")
	rootCmd.PersistentFlags().StringP("project", "p", "", "Override default project ID")
	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolP("verbose", "V", false, "Enable verbose output for debugging")

	// Bind flags to viper
	_ = viper.BindPFlag("account", rootCmd.PersistentFlags().Lookup("account"))
	_ = viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

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
	rootCmd.AddCommand(profile.NewProfileCmd(f))

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
