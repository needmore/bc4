package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui"
)

func newListCmd() *cobra.Command {
	var jsonOutput bool
	var accountID string
	var formatStr string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all projects",
		Long:    `List all projects in your Basecamp account. Use 'project select' for interactive selection.`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if we have auth
			if cfg.ClientID == "" || cfg.ClientSecret == "" {
				return fmt.Errorf("not authenticated. Run 'bc4' to set up authentication")
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

			// Use specified account or default
			if accountID == "" {
				accountID = authClient.GetDefaultAccount()
			}

			if accountID == "" {
				return fmt.Errorf("no account specified and no default account set")
			}

			// Get token
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			apiClient := api.NewClient(accountID, token.AccessToken)

			// Fetch projects
			projects, err := apiClient.GetProjects(context.Background())
			if err != nil {
				return fmt.Errorf("failed to fetch projects: %w", err)
			}

			// Sort projects alphabetically
			sortProjectsByName(projects)

			// Get default project ID
			defaultProjectID := cfg.DefaultProject
			if defaultProjectID == "" && cfg.Accounts != nil {
				if acc, ok := cfg.Accounts[accountID]; ok {
					defaultProjectID = acc.DefaultProject
				}
			}

			// Output JSON if requested
			if jsonOutput {
				return outputJSON(projects)
			}

			// Check if there are any projects
			if len(projects) == 0 {
				fmt.Println("No projects found.")
				return nil
			}

			// Parse output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			// Create output config
			config := ui.NewOutputConfig(os.Stdout)
			config.Format = format
			config.NoHeaders = true // We'll add custom headers

			// Create table writer
			tw := ui.NewTableWriter(config)

			// Add rows (headers are implicit from the data)
			for _, project := range projects {
				row := []string{
					project.Name,
					strconv.FormatInt(project.ID, 10),
					project.Description,
				}

				// Add a marker for the default project
				if strconv.FormatInt(project.ID, 10) == defaultProjectID {
					if config.Format == ui.OutputFormatTable && ui.IsTerminal(os.Stdout) && !config.NoColor {
						// Add a subtle indicator for the default project
						row[0] = "→ " + row[0]
					}
				}

				tw.AddRow(row)
			}

			return tw.Render()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (deprecated, use --format=json)")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or tsv")

	return cmd
}

func sortProjectsByName(projects []api.Project) {
	// Using bubble sort for simplicity, could use sort.Slice
	for i := 0; i < len(projects); i++ {
		for j := i + 1; j < len(projects); j++ {
			if strings.ToLower(projects[i].Name) > strings.ToLower(projects[j].Name) {
				projects[i], projects[j] = projects[j], projects[i]
			}
		}
	}
}

func outputJSON(projects []api.Project) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(projects)
}
