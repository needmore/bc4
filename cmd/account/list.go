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

			// Create output config
			config := ui.NewOutputConfig(os.Stdout)
			config.Format = format
			config.NoHeaders = true // We'll add custom formatting

			// Create table writer
			tw := ui.NewTableWriter(config)

			// Add rows
			for _, acc := range accountList {
				row := []string{
					acc.Name,
					acc.ID,
				}

				// Add a marker for the default account
				if acc.Default {
					if config.Format == ui.OutputFormatTable && ui.IsTerminal(os.Stdout) && !config.NoColor {
						// Add a subtle indicator for the default account
						row[0] = "→ " + row[0]
					} else if config.Format != ui.OutputFormatTable {
						// For non-table formats, add a third column
						row = append(row, "default")
					}
				} else if config.Format != ui.OutputFormatTable {
					// Add empty third column for consistency
					row = append(row, "")
				}

				tw.AddRow(row)
			}

			return tw.Render()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (deprecated, use --format=json)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or tsv")

	return cmd
}
