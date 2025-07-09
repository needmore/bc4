package account

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui"
)

func newListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all accounts",
		Long:  `List all authenticated Basecamp accounts. Use 'account select' for interactive selection.`,
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
			type accountInfo struct {
				ID      string
				Name    string
				Default bool
			}

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

			// Output JSON if requested
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(accountList)
			}

			// Get terminal width for responsive columns
			termWidth := ui.GetTerminalWidth()

			// Calculate column widths
			// The table adds borders and padding, so we need to account for that
			// Each column separator is 3 chars (space + | + space)
			// Plus 2 for the outer borders
			tableOverhead := 3 + 2
			availableWidth := termWidth - tableOverhead
			
			nameWidth := 50
			idWidth := 15
			
			// If we have more space than needed, use it
			if availableWidth > nameWidth + idWidth {
				extraSpace := availableWidth - nameWidth - idWidth
				nameWidth += extraSpace / 2
				idWidth += extraSpace / 2
			} else if availableWidth < nameWidth + idWidth {
				// Scale down proportionally
				ratio := float64(availableWidth) / float64(nameWidth + idWidth)
				nameWidth = int(float64(nameWidth) * ratio)
				idWidth = int(float64(idWidth) * ratio)
				if nameWidth < 20 {
					nameWidth = 20
				}
				if idWidth < 8 {
					idWidth = 8
				}
			}

			// Create table
			columns := []table.Column{
				{Title: "", Width: nameWidth},
				{Title: "", Width: idWidth},
			}

			rows := []table.Row{}
			defaultIndex := 0
			for i, acc := range accountList {
				name := ui.TruncateString(acc.Name, nameWidth-3)
				
				// Track which row is the default
				if acc.Default {
					defaultIndex = i
				}
				
				rows = append(rows, table.Row{
					name,
					acc.ID,
				})
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithHeight(len(rows)+2), // Include space for header and borders
			)

			// Style the table with subtle list highlighting
			t = ui.StyleTableForList(t)
			
			// Set cursor to the default account
			t.SetCursor(defaultIndex)

			// Print the table, skipping the empty header row
			tableView := t.View()
			lines := strings.Split(tableView, "\n")
			if len(lines) > 2 {
				// Skip first line (top border) and second line (empty header)
				// Keep the top border but skip the header
				result := lines[0] + "\n" + strings.Join(lines[2:], "\n")
				fmt.Println(ui.BaseTableStyle.Render(result))
			} else {
				fmt.Println(ui.BaseTableStyle.Render(tableView))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}