package project

import (
	"fmt"
	"io"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
)

type projectsLoadedMsg struct {
	projects []api.Project
	err      error
}

type selectModel struct {
	list      list.Model
	projects  []api.Project
	spinner   spinner.Model
	loading   bool
	err       error
	width     int
	height    int
	accountID string
	factory   *factory.Factory
}

func (m selectModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadProjects(),
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
			listWidth := min(m.width-20, 70)
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
			if selected, ok := m.list.SelectedItem().(projectItem); ok {
				for _, p := range m.projects {
					if strconv.FormatInt(p.ID, 10) == selected.id {
						return m, tea.Sequence(
							m.saveDefaultProject(p),
							tea.Quit,
						)
					}
				}
			}
		}

	case projectsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.loading = false
			return m, tea.Quit
		}

		m.projects = msg.projects
		m.loading = false

		// Sort projects alphabetically
		sortProjectsByName(m.projects)

		// Create list items
		items := make([]list.Item, 0, len(m.projects))
		for _, project := range m.projects {
			items = append(items, projectItem{
				id:   strconv.FormatInt(project.ID, 10),
				name: project.Name,
				desc: project.Description,
			})
		}

		// Calculate list dimensions
		listHeight := min(m.height-10, len(items)*3+6)
		listWidth := min(m.width-20, 70)

		// Create list with custom delegate
		m.list = list.New(items, itemDelegate{}, listWidth, listHeight)
		m.list.Title = "Select Project"
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
		return fmt.Sprintf("\n  %s Loading projects...\n\n", m.spinner.View())
	}

	if len(m.projects) == 0 {
		return "\n  No projects found.\n\n"
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

func (m *selectModel) loadProjects() tea.Cmd {
	return func() tea.Msg {
		// Create API client through factory
		apiClient, err := m.factory.ApiClient()
		if err != nil {
			return projectsLoadedMsg{err: err}
		}
		projectOps := apiClient.Projects()

		// Fetch projects
		projects, err := projectOps.GetProjects(m.factory.Context())
		if err != nil {
			return projectsLoadedMsg{err: err}
		}

		return projectsLoadedMsg{projects: projects}
	}
}

func (m *selectModel) saveDefaultProject(project api.Project) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return nil
		}

		cfg.DefaultProject = fmt.Sprintf("%d", project.ID)

		if cfg.Accounts == nil {
			cfg.Accounts = make(map[string]config.AccountConfig)
		}

		// Update account-specific default project
		accountCfg := cfg.Accounts[m.accountID]
		accountCfg.DefaultProject = fmt.Sprintf("%d", project.ID)
		// Preserve the name if it exists
		if accountCfg.Name == "" {
			// Get the account name from auth
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)
			if token, err := authClient.GetToken(m.accountID); err == nil {
				accountCfg.Name = token.AccountName
			}
		}
		cfg.Accounts[m.accountID] = accountCfg

		_ = config.Save(cfg)

		fmt.Printf("\nDefault project set to: %s (ID: %d)\n", project.Name, project.ID)
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
)

// projectItem implements list.Item
type projectItem struct {
	id   string
	name string
	desc string
}

func (i projectItem) FilterValue() string { return i.name }
func (i projectItem) Title() string       { return i.name }
func (i projectItem) Description() string { return i.desc }

// Custom item delegate for cleaner rendering
type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 2 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(projectItem)
	if !ok {
		return
	}

	// Render name
	name := i.name
	if index == m.Index() {
		_, _ = fmt.Fprintln(w, selectedItemStyle.Render("→ "+name))
	} else {
		_, _ = fmt.Fprintln(w, normalItemStyle.Render("  "+name))
	}

	// Render description on second line
	if i.desc != "" {
		desc := i.desc
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(4)
		_, _ = fmt.Fprintln(w, descStyle.Render(desc))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func newSelectCmd(f *factory.Factory) *cobra.Command {
	var accountID string

	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select default project",
		Long:  `Interactively select a default project for bc4 commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get auth through factory
			authClient, err := f.AuthClient()
			if err != nil {
				return err
			}

			// Use specified account or default
			if accountID == "" {
				accountID = authClient.GetDefaultAccount()
			}

			if accountID == "" {
				return fmt.Errorf("no account specified and no default account set")
			}

			// Create spinner
			s := spinner.New()
			s.Spinner = spinner.Dot
			s.Style = ui.SelectedItemStyle

			// If accountID was specified, use a new factory with that account
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Create model
			m := selectModel{
				spinner:   s,
				loading:   true,
				accountID: accountID,
				factory:   f,
			}

			// Run the interactive selector
			if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
				return fmt.Errorf("error running selector: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")

	return cmd
}
