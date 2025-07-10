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
	"github.com/needmore/bc4/internal/ui/tableprinter"
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

			// Check output format for non-table output
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			if format == ui.OutputFormatJSON {
				return outputJSON(projects)
			}

			// Check if there are any projects
			if len(projects) == 0 {
				fmt.Println("No projects found.")
				return nil
			}

			// Create new GitHub CLI-style table
			table := tableprinter.New(os.Stdout)

			// Add headers dynamically based on TTY mode (like GitHub CLI)
			if table.IsTTY() {
				table.AddHeader("ID", "NAME", "DESCRIPTION", "UPDATED")
			} else {
				// Add STATE column for non-TTY mode (machine readable)
				table.AddHeader("ID", "NAME", "DESCRIPTION", "STATE", "UPDATED")
			}

			// Add projects to table
			for _, project := range projects {
				// For now, assume all projects are active (no status field in API yet)
				state := "active"

				// Add ID field with color based on state and default indicator
				projectID := strconv.FormatInt(project.ID, 10)
				if projectID == defaultProjectID {
					// Mark default project with special color/indicator
					if table.IsTTY() {
						table.AddIDField(projectID+"*", state) // Add asterisk for default
					} else {
						table.AddIDField(projectID, state)
					}
				} else {
					table.AddIDField(projectID, state)
				}

				// Add project name with appropriate coloring
				table.AddProjectField(project.Name, state)

				// Add description with muted color
				cs := table.GetColorScheme()
				table.AddField(project.Description, cs.Muted)

				// Add STATE column only for non-TTY
				if !table.IsTTY() {
					table.AddField(state)
				}

				// Add updated time - use UpdatedAt if available, otherwise created
				timeStr := project.UpdatedAt
				if timeStr == "" {
					timeStr = project.CreatedAt
				}
				table.AddField(timeStr, cs.Muted)

				table.EndRow()
			}

			return table.Render()
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
