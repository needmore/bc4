package project

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
)

var (
	selectBaseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
)

type projectsLoadedMsg struct {
	projects []api.Project
	err      error
}

type selectModel struct {
	table        table.Model
	projects     []api.Project
	spinner      spinner.Model
	loading      bool
	err          error
	width        int
	height       int
	accountID    string
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
		// Only update table dimensions if it's been initialized
		if m.table.Cursor() >= 0 {
			m.table.SetWidth(msg.Width)
			// Calculate table height based on window height
			tableHeight := msg.Height - 6
			if tableHeight < 10 {
				tableHeight = 10
			}
			// Don't make it taller than needed
			rowCount := len(m.table.Rows())
			if tableHeight > rowCount+2 {
				tableHeight = rowCount + 2
			}
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
				// Get the project ID from the selected row
				projectID := selectedRow[1]
				for _, p := range m.projects {
					if strconv.FormatInt(p.ID, 10) == projectID {
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

		// Create table columns
		columns := []table.Column{
			{Title: "Name", Width: 40},
			{Title: "ID", Width: 10},
			{Title: "Description", Width: 50},
		}

		// Create rows
		rows := []table.Row{}
		for _, project := range m.projects {
			desc := project.Description
			// Don't fall back to Purpose - just leave blank if no description
			// Truncate description if too long
			if len(desc) > 47 {
				desc = desc[:44] + "..."
			}
			
			rows = append(rows, table.Row{
				project.Name,
				strconv.FormatInt(project.ID, 10),
				desc,
			})
		}

		// Calculate table height based on window height
		// Leave room for borders, header, and help text
		tableHeight := m.height - 6
		if tableHeight < 10 {
			tableHeight = 10 // Minimum height
		}
		// Don't make it taller than needed
		if tableHeight > len(rows)+2 {
			tableHeight = len(rows) + 2
		}

		// Create table
		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(tableHeight),
		)

		// Style the table
		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

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
		return fmt.Sprintf("\n  %s Loading projects...\n\n", m.spinner.View())
	}

	if len(m.projects) == 0 {
		return "\n  No projects found.\n\n"
	}

	return selectBaseStyle.Render(m.table.View()) + "\n" + helpStyle.Render("↑/↓: Navigate • Enter: Select • q/Esc: Cancel")
}

func (m *selectModel) loadProjects() tea.Cmd {
	return func() tea.Msg {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return projectsLoadedMsg{err: err}
		}

		// Create auth client
		authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

		// Get token
		token, err := authClient.GetToken(m.accountID)
		if err != nil {
			return projectsLoadedMsg{err: err}
		}

		// Create API client
		apiClient := api.NewClient(m.accountID, token.AccessToken)

		// Fetch projects
		projects, err := apiClient.GetProjects(context.Background())
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

		config.Save(cfg)

		fmt.Printf("\nDefault project set to: %s (ID: %d)\n", project.Name, project.ID)
		return nil
	}
}


func newSelectCmd() *cobra.Command {
	var accountID string

	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select default project",
		Long:  `Interactively select a default project for bc4 commands.`,
		Aliases: []string{"set-default", "default"},
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

			// Create spinner
			s := spinner.New()
			s.Spinner = spinner.Dot
			s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

			// Create model
			m := selectModel{
				spinner:   s,
				loading:   true,
				accountID: accountID,
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