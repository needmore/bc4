package account

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	list     list.Model
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
		// Update list dimensions if it's been initialized
		if m.list.Items() != nil {
			listHeight := min(m.height-10, len(m.list.Items())*3+6)
			listWidth := min(m.width-20, 60)
			m.list.SetWidth(listWidth)
			m.list.SetHeight(listHeight)
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
			if selected, ok := m.list.SelectedItem().(accountListItem); ok {
				return m, tea.Sequence(
					m.setDefaultAccount(selected.id, selected.name),
					tea.Quit,
				)
			}
		}

	case accountsLoadedMsg:
		m.accounts = msg.accounts
		m.loading = false

		if len(m.accounts) == 0 {
			m.err = fmt.Errorf("no accounts found")
			return m, tea.Quit
		}

		// Create list items
		items := make([]list.Item, 0, len(m.accounts))
		for _, acc := range m.accounts {
			items = append(items, accountListItem{
				id:      acc.id,
				name:    acc.name,
				current: acc.current,
			})
		}

		// Calculate list dimensions
		listHeight := min(m.height-10, len(items)*2+6)
		listWidth := min(m.width-20, 60)

		// Create list with custom delegate
		m.list = list.New(items, accountDelegate{}, listWidth, listHeight)
		m.list.Title = "Select Account"
		m.list.SetShowStatusBar(false)
		m.list.SetFilteringEnabled(true)
		m.list.SetShowHelp(false)
		m.list.Styles.Title = titleStyle
		m.list.Styles.TitleBar = lipgloss.NewStyle()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if !m.loading {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
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

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.list.View(),
		"",
		helpStyle.Render("↑/↓: Navigate • Enter: Select • /: Filter • Esc: Cancel"),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
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

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	currentMarkerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true)
)

// accountListItem implements list.Item
type accountListItem struct {
	id      string
	name    string
	current bool
}

func (i accountListItem) FilterValue() string { return i.name }
func (i accountListItem) Title() string       { return i.name }
func (i accountListItem) Description() string { return fmt.Sprintf("ID: %s", i.id) }

// Custom delegate for account items
type accountDelegate struct{}

func (d accountDelegate) Height() int                               { return 1 }
func (d accountDelegate) Spacing() int                              { return 0 }
func (d accountDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d accountDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(accountListItem)
	if !ok {
		return
	}

	// Build the display string
	str := i.name
	if i.current {
		str = str + " " + currentMarkerStyle.Render("✓")
	}

	if index == m.Index() {
		_, _ = fmt.Fprint(w, selectedItemStyle.Render("→ "+str))
	} else {
		_, _ = fmt.Fprint(w, normalItemStyle.Render("  "+str))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
