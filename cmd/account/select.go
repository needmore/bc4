package account

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
)

type accountsLoadedMsg struct {
	accounts []accountItem
}

type accountItem struct {
	id      string
	name    string
	current bool
}

type selectModel struct {
	table    table.Model
	accounts []accountItem
	spinner  spinner.Model
	loading  bool
	err      error
	width    int
	height   int
	factory  *factory.Factory
}

func (m selectModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadAccounts(),
	)
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Only update table dimensions if it's been initialized
		if m.table.Cursor() >= 0 {
			m.table.SetWidth(msg.Width)
			tableHeight := ui.CalculateTableHeight(msg.Height, len(m.table.Rows()))
			m.table.SetHeight(tableHeight)
		}
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			selectedRow := m.table.SelectedRow()
			if len(selectedRow) >= 2 {
				// Get the account ID from the selected row
				accountID := selectedRow[1]
				for _, acc := range m.accounts {
					if acc.id == accountID {
						return m, tea.Sequence(
							m.setDefaultAccount(accountID, acc.name),
							tea.Quit,
						)
					}
				}
			}
		}

	case accountsLoadedMsg:
		m.accounts = msg.accounts
		m.loading = false

		if len(m.accounts) == 0 {
			m.err = fmt.Errorf("no accounts found")
			return m, tea.Quit
		}

		// Create table columns
		columns := []table.Column{
			{Title: "", Width: 40},
			{Title: "", Width: 10},
			{Title: "", Width: 10},
		}

		// Create rows
		rows := []table.Row{}
		for _, acc := range m.accounts {
			defaultStr := ""
			if acc.current {
				defaultStr = "✓"
			}

			rows = append(rows, table.Row{
				acc.name,
				acc.id,
				defaultStr,
			})
		}

		// Create table
		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(ui.CalculateTableHeight(m.height, len(rows))),
		)

		// Apply common table styling
		t = ui.StyleTable(t)

		m.table = t
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if !m.loading {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m selectModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n", m.err)
	}

	if m.loading {
		return fmt.Sprintf("\n  %s Loading accounts...\n\n", m.spinner.View())
	}

	if len(m.accounts) == 0 {
		return "\n  No accounts found.\n\n"
	}

	return ui.BaseTableStyle.Render(m.table.View()) + "\n" + ui.HelpStyle.Render("↑/↓: Navigate • Enter: Select • q/Esc: Cancel")
}

func (m *selectModel) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return accountsLoadedMsg{}
		}

		// Create auth client
		authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

		// Get all accounts
		accounts := authClient.GetAccounts()
		defaultAccount := authClient.GetDefaultAccount()

		// Convert to sorted slice
		var accountList []accountItem
		for id, acc := range accounts {
			accountList = append(accountList, accountItem{
				id:      id,
				name:    acc.AccountName,
				current: id == defaultAccount,
			})
		}

		// Sort accounts by name
		sort.Slice(accountList, func(i, j int) bool {
			return strings.ToLower(accountList[i].name) < strings.ToLower(accountList[j].name)
		})

		return accountsLoadedMsg{accounts: accountList}
	}
}

func (m *selectModel) setDefaultAccount(accountID, accountName string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return nil
		}

		// Create auth client and set default
		authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

		// Check if we're changing accounts
		oldDefaultAccount := authClient.GetDefaultAccount()
		changingAccounts := oldDefaultAccount != "" && oldDefaultAccount != accountID

		if err := authClient.SetDefaultAccount(accountID); err != nil {
			return nil
		}

		// Update config
		cfg.DefaultAccount = accountID

		// Clear default project if changing accounts
		if changingAccounts {
			cfg.DefaultProject = ""
			// Also clear the account-specific default project
			if cfg.Accounts != nil {
				for accID, accConfig := range cfg.Accounts {
					if accID == accountID {
						accConfig.DefaultProject = ""
						cfg.Accounts[accID] = accConfig
					}
				}
			}
		}

		// Save config
		_ = config.Save(cfg)

		fmt.Printf("\nDefault account set to: %s (ID: %s)\n", accountName, accountID)
		if changingAccounts {
			fmt.Println("Note: Default project has been cleared since you changed accounts.")
		}
		return nil
	}
}

func newSelectCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select default account",
		Long:  `Interactively select a default account for bc4 commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create spinner
			s := spinner.New()
			s.Spinner = spinner.Dot
			s.Style = ui.SelectedItemStyle

			// Create model
			m := selectModel{
				spinner: s,
				loading: true,
				factory: f,
			}

			// Run the interactive selector
			if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
				return fmt.Errorf("error running selector: %w", err)
			}

			return nil
		},
	}

	return cmd
}
