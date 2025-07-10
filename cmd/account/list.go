package account

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui"
	"github.com/needmore/bc4/internal/ui/tableprinter"
)

type accountInfo struct {
	ID      string
	Name    string
	Default bool
}

func newListCmd() *cobra.Command {
	var jsonOutput bool
	var formatStr string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all accounts",
		Long:    `List all authenticated Basecamp accounts. Use 'account select' for interactive selection.`,
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

			// Get all accounts
			accounts := authClient.GetAccounts()

			// Check if there are any accounts
			if len(accounts) == 0 {
				fmt.Println("No authenticated accounts found.")
				return nil
			}

			// Convert to sorted slice
			var accountList []accountInfo
			defaultAccount := authClient.GetDefaultAccount()

			for id, acc := range accounts {
				accountList = append(accountList, accountInfo{
					ID:      id,
					Name:    acc.AccountName,
					Default: id == defaultAccount,
				})
			}

			// Sort accounts by name
			sort.Slice(accountList, func(i, j int) bool {
				return strings.ToLower(accountList[i].Name) < strings.ToLower(accountList[j].Name)
			})

			// Parse output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			// Handle legacy JSON flag
			if jsonOutput {
				format = ui.OutputFormatJSON
			}

			// Handle JSON output directly
			if format == ui.OutputFormatJSON {
				// TODO: Implement JSON output for accounts
				return fmt.Errorf("JSON output not yet implemented for accounts")
			}

			// Create new GitHub CLI-style table
			table := tableprinter.New(os.Stdout)

			// Add headers dynamically based on TTY mode (like GitHub CLI)
			if table.IsTTY() {
				table.AddHeader("ID", "NAME", "UPDATED")
			} else {
				// Add STATE column for non-TTY mode (machine readable)
				table.AddHeader("ID", "NAME", "STATE", "UPDATED")
			}

			// Add accounts to table
			for _, acc := range accountList {
				// Determine account state
				state := "active"
				if acc.Default {
					state = "default"
				}

				// Add ID field with default indicator
				if acc.Default {
					if table.IsTTY() {
						table.AddIDField(acc.ID+"*", state) // Add asterisk for default
					} else {
						table.AddIDField(acc.ID, state)
					}
				} else {
					table.AddIDField(acc.ID, state)
				}

				// Add account name with appropriate coloring
				cs := table.GetColorScheme()
				table.AddField(acc.Name, cs.AccountName)

				// Add STATE column only for non-TTY
				if !table.IsTTY() {
					table.AddField(state)
				}

				// Add updated time placeholder
				table.AddField("N/A", cs.Muted)

				table.EndRow()
			}

			return table.Render()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (deprecated, use --format=json)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or tsv")

	return cmd
}
