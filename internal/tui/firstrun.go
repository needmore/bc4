package tui

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)

type step int

const (
	stepWelcome step = iota
	stepClientID
	stepClientSecret
	stepAuthenticate
	stepSelectAccount
	stepSelectProject
	stepComplete
)

type authenticateMsg struct {
	token *auth.AccountToken
	err   error
}

type accountsLoadedMsg struct {
	accounts map[string]auth.AccountToken
}

type projectsLoadedMsg struct {
	projects []api.Project
	err      error
}

// FirstRunModel represents the first-run wizard
type FirstRunModel struct {
	currentStep     step
	clientID        textinput.Model
	clientSecret    textinput.Model
	authClient      *auth.Client
	token           *auth.AccountToken
	accounts        map[string]auth.AccountToken
	accountList     list.Model
	selectedAccount string
	projects        []api.Project
	projectList     list.Model
	spinner         spinner.Model
	err             error
	width           int
	height          int
}

// NewFirstRunModel creates a new first-run wizard
func NewFirstRunModel() FirstRunModel {
	// Client ID input
	clientID := textinput.New()
	clientID.Placeholder = "Your Basecamp OAuth Client ID"
	clientID.Focus()
	clientID.PromptStyle = focusedStyle
	clientID.TextStyle = focusedStyle
	clientID.CharLimit = 64

	// Client Secret input
	clientSecret := textinput.New()
	clientSecret.Placeholder = "Your Basecamp OAuth Client Secret"
	clientSecret.PromptStyle = blurredStyle
	clientSecret.TextStyle = blurredStyle
	clientSecret.CharLimit = 64
	clientSecret.EchoMode = textinput.EchoPassword

	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = focusedStyle

	return FirstRunModel{
		currentStep:  stepWelcome,
		clientID:     clientID,
		clientSecret: clientSecret,
		spinner:      s,
	}
}

// Init initializes the model
func (m FirstRunModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

// Update handles messages
func (m FirstRunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			// Allow ESC to skip project selection
			if m.currentStep == stepSelectProject {
				// Save config without project
				cfg, _ := config.Load()
				if cfg == nil {
					cfg = &config.Config{
						Accounts: make(map[string]config.AccountConfig),
					}
				}
				// Ensure OAuth credentials are set from user input
				if cfg.ClientID == "" {
					cfg.ClientID = m.clientID.Value()
				}
				if cfg.ClientSecret == "" {
					cfg.ClientSecret = m.clientSecret.Value()
				}
				cfg.DefaultAccount = m.selectedAccount
				if err := config.Save(cfg); err != nil {
					m.err = fmt.Errorf("failed to save config: %w", err)
					return m, nil
				}
				m.currentStep = stepComplete
				return m, nil
			}
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyTab, tea.KeyShiftTab:
			switch m.currentStep {
			case stepClientID:
				m.clientID.Blur()
				m.clientSecret.Focus()
				m.currentStep = stepClientSecret
			case stepClientSecret:
				m.clientSecret.Blur()
				m.clientID.Focus()
				m.currentStep = stepClientID
			}
		}

	case authenticateMsg:
		if msg.err != nil {
			m.err = msg.err
			m.currentStep = stepClientSecret
			return m, nil
		}
		m.token = msg.token
		// Store the auth client for later use
		m.authClient = auth.NewClient(m.clientID.Value(), m.clientSecret.Value())
		// Load accounts for selection
		m.currentStep = stepSelectAccount
		return m, m.loadAccounts()

	case accountsLoadedMsg:
		m.accounts = msg.accounts

		if len(m.accounts) == 0 {
			m.err = fmt.Errorf("no accounts found")
			m.currentStep = stepAuthenticate
			return m, nil
		}

		// If only one account, skip to project selection
		if len(m.accounts) == 1 {
			for accountID := range m.accounts {
				_ = m.authClient.SetDefaultAccount(accountID)
				m.selectedAccount = accountID
				// Load projects for this account
				m.currentStep = stepSelectProject
				return m, m.loadProjects(accountID)
			}
		}

		// Multiple accounts - show selection
		// First collect accounts into a slice we can sort
		var accountList []auth.AccountToken
		for _, account := range m.accounts {
			accountList = append(accountList, account)
		}

		// Sort accounts alphabetically by name
		sort.Slice(accountList, func(i, j int) bool {
			return strings.ToLower(accountList[i].AccountName) < strings.ToLower(accountList[j].AccountName)
		})

		// Create items from sorted list
		items := make([]list.Item, 0, len(accountList))
		for _, account := range accountList {
			items = append(items, accountItem{
				id:   account.AccountID,
				name: account.AccountName,
			})
		}

		// Use our custom delegate for cleaner rendering
		m.accountList = list.New(items, itemDelegate{}, 50, min(10, len(items)+2))
		m.accountList.Title = "Select Default Account"
		m.accountList.SetShowStatusBar(false)
		m.accountList.SetFilteringEnabled(false)
		m.accountList.SetShowHelp(false)
		m.accountList.Styles.Title = titleStyle
		m.accountList.Styles.TitleBar = lipgloss.NewStyle()
		return m, nil

	case projectsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.currentStep = stepSelectAccount
			return m, nil
		}

		m.projects = msg.projects

		// Skip project selection if no projects
		if len(m.projects) == 0 {
			m.currentStep = stepComplete
			return m, nil
		}

		// Sort projects alphabetically by name
		sort.Slice(m.projects, func(i, j int) bool {
			return strings.ToLower(m.projects[i].Name) < strings.ToLower(m.projects[j].Name)
		})

		// Create project list
		items := make([]list.Item, 0, len(m.projects))
		for _, project := range m.projects {
			items = append(items, projectItem{
				id:   fmt.Sprintf("%d", project.ID),
				name: project.Name,
				desc: project.Description,
			})
		}

		// Use our custom delegate for cleaner rendering
		m.projectList = list.New(items, itemDelegate{}, 60, min(15, len(items)+2))
		m.projectList.Title = "Select Default Project (Optional)"
		m.projectList.SetShowStatusBar(false)
		m.projectList.SetFilteringEnabled(true)
		m.projectList.SetShowHelp(false)
		m.projectList.Styles.Title = titleStyle
		m.projectList.Styles.TitleBar = lipgloss.NewStyle()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Update text inputs
	switch m.currentStep {
	case stepClientID:
		var cmd tea.Cmd
		m.clientID, cmd = m.clientID.Update(msg)
		return m, cmd

	case stepClientSecret:
		var cmd tea.Cmd
		m.clientSecret, cmd = m.clientSecret.Update(msg)
		return m, cmd

	case stepSelectAccount:
		if len(m.accountList.Items()) > 0 {
			var cmd tea.Cmd
			m.accountList, cmd = m.accountList.Update(msg)
			return m, cmd
		}

	case stepSelectProject:
		if len(m.projectList.Items()) > 0 {
			var cmd tea.Cmd
			m.projectList, cmd = m.projectList.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the UI
func (m FirstRunModel) View() string {
	switch m.currentStep {
	case stepWelcome:
		return m.viewWelcome()
	case stepClientID:
		return m.viewClientID()
	case stepClientSecret:
		return m.viewClientSecret()
	case stepAuthenticate:
		return m.viewAuthenticate()
	case stepSelectAccount:
		return m.viewSelectAccount()
	case stepSelectProject:
		return m.viewSelectProject()
	case stepComplete:
		return m.viewComplete()
	default:
		return ""
	}
}

func (m FirstRunModel) viewWelcome() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Welcome to bc4!"),
		subtitleStyle.Render("The Basecamp command-line interface"),
		"",
		"This wizard will help you:",
		"• Create a Basecamp OAuth application",
		"• Authenticate with your Basecamp account",
		"• Configure your default settings",
		"",
		"Before we begin, you'll need to create an OAuth app:",
		"1. Visit https://launchpad.37signals.com/integrations",
		"2. Click 'Register one now'",
		"3. Set Redirect URI to: http://localhost:8888/callback",
		"4. Save your Client ID and Client Secret",
		"",
		helpStyle.Render("Press Enter to continue..."),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m FirstRunModel) viewClientID() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("OAuth Setup"),
		subtitleStyle.Render("Enter your Basecamp OAuth credentials"),
		"",
		"Client ID:",
		m.clientID.View(),
		"",
		helpStyle.Render("Tab to switch fields • Enter to continue"),
	)

	if m.err != nil && m.currentStep == stepClientID {
		content = lipgloss.JoinVertical(lipgloss.Left,
			content,
			"",
			errorStyle.Render("Error: "+m.err.Error()),
		)
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m FirstRunModel) viewClientSecret() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("OAuth Setup"),
		subtitleStyle.Render("Enter your Basecamp OAuth credentials"),
		"",
		"Client ID:",
		blurredStyle.Render(m.clientID.Value()),
		"",
		"Client Secret:",
		m.clientSecret.View(),
		"",
		helpStyle.Render("Tab to switch fields • Enter to continue"),
	)

	if m.err != nil && m.currentStep == stepClientSecret {
		content = lipgloss.JoinVertical(lipgloss.Left,
			content,
			"",
			errorStyle.Render("Error: "+m.err.Error()),
		)
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m FirstRunModel) viewAuthenticate() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Authenticating with Basecamp"),
		"",
		m.spinner.View()+" Opening your browser...",
		"",
		subtitleStyle.Render("Please authorize bc4 in your browser"),
		"",
		helpStyle.Render("Waiting for authorization..."),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m FirstRunModel) viewSelectAccount() string {
	if m.accountList.Items() == nil || len(m.accountList.Items()) == 0 {
		// Still loading
		content := lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Loading accounts..."),
			"",
			m.spinner.View(),
		)
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			content,
		)
	}

	// Update list dimensions to fit window
	listHeight := min(m.height-10, len(m.accountList.Items())*3+6)
	listWidth := min(m.width-20, 60)
	m.accountList.SetWidth(listWidth)
	m.accountList.SetHeight(listHeight)

	// Show account list
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Select Default Account"),
		"",
		m.accountList.View(),
		"",
		helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel"),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m FirstRunModel) viewSelectProject() string {
	if m.projectList.Items() == nil || len(m.projectList.Items()) == 0 {
		// Still loading or no projects
		content := lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Loading projects..."),
			subtitleStyle.Render("Fetching all projects from your account"),
			"",
			m.spinner.View()+" Please wait, this may take a moment...",
		)
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			content,
		)
	}

	// Update list dimensions to fit window
	listHeight := min(m.height-12, len(m.projectList.Items())*4+6)
	listWidth := min(m.width-20, 70)
	m.projectList.SetWidth(listWidth)
	m.projectList.SetHeight(listHeight)

	// Show project list
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Select Default Project"),
		subtitleStyle.Render("Optional - you can skip this step"),
		"",
		m.projectList.View(),
		"",
		helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Skip"),
	)

	if m.err != nil {
		content = lipgloss.JoinVertical(lipgloss.Left,
			content,
			"",
			errorStyle.Render("Error: "+m.err.Error()),
		)
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m FirstRunModel) viewComplete() string {
	accountName := "your account"
	if m.token != nil {
		accountName = m.token.AccountName
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Setup Complete! 🎉"),
		"",
		successStyle.Render(fmt.Sprintf("Successfully authenticated with %s", accountName)),
		"",
		"You can now use bc4 to:",
		"• List and manage projects",
		"• Create and complete todos",
		"• Post messages and campfire updates",
		"• Manage cards on kanban boards",
		"",
		"Try these commands:",
		focusedStyle.Render("  bc4 project list"),
		focusedStyle.Render("  bc4 todo create"),
		focusedStyle.Render("  bc4 auth status"),
		"",
		helpStyle.Render("Press Enter to exit..."),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m FirstRunModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentStep {
	case stepWelcome:
		m.currentStep = stepClientID
		return m, nil

	case stepClientID:
		if strings.TrimSpace(m.clientID.Value()) == "" {
			m.err = fmt.Errorf("client ID is required")
			return m, nil
		}
		m.currentStep = stepClientSecret
		m.clientID.Blur()
		m.clientSecret.Focus()
		return m, nil

	case stepClientSecret:
		if strings.TrimSpace(m.clientSecret.Value()) == "" {
			m.err = fmt.Errorf("client secret is required")
			return m, nil
		}
		m.currentStep = stepAuthenticate
		return m, m.authenticate()

	case stepSelectAccount:
		if selected, ok := m.accountList.SelectedItem().(accountItem); ok {
			// Set default account
			_ = m.authClient.SetDefaultAccount(selected.id)
			m.selectedAccount = selected.id
			// Load projects for selected account
			m.currentStep = stepSelectProject
			return m, m.loadProjects(selected.id)
		}
		return m, nil

	case stepSelectProject:
		cfg, _ := config.Load()
		if cfg == nil {
			cfg = &config.Config{
				Accounts: make(map[string]config.AccountConfig),
			}
		}
		// Ensure OAuth credentials are set from user input
		if cfg.ClientID == "" {
			cfg.ClientID = m.clientID.Value()
		}
		if cfg.ClientSecret == "" {
			cfg.ClientSecret = m.clientSecret.Value()
		}

		// Save default account
		cfg.DefaultAccount = m.selectedAccount

		// Save selected project if any
		if selected, ok := m.projectList.SelectedItem().(projectItem); ok {
			cfg.DefaultProject = selected.id
			if cfg.Accounts == nil {
				cfg.Accounts = make(map[string]config.AccountConfig)
			}
			cfg.Accounts[m.selectedAccount] = config.AccountConfig{
				Name:           m.accounts[m.selectedAccount].AccountName,
				DefaultProject: selected.id,
			}
		}

		if err := config.Save(cfg); err != nil {
			m.err = fmt.Errorf("failed to save config: %w", err)
			return m, nil
		}
		m.currentStep = stepComplete
		return m, nil

	case stepComplete:
		// Config has already been saved in previous steps
		return m, tea.Quit
	}

	return m, nil
}

func (m *FirstRunModel) authenticate() tea.Cmd {
	return func() tea.Msg {
		m.authClient = auth.NewClient(m.clientID.Value(), m.clientSecret.Value())
		token, err := m.authClient.Login(context.Background())
		return authenticateMsg{token: token, err: err}
	}
}

func (m *FirstRunModel) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		accounts := m.authClient.GetAccounts()
		return accountsLoadedMsg{accounts: accounts}
	}
}

func (m *FirstRunModel) loadProjects(accountID string) tea.Cmd {
	return func() tea.Msg {
		// Get the account token
		token, err := m.authClient.GetToken(accountID)
		if err != nil {
			return projectsLoadedMsg{err: err}
		}

		// Create modular API client
		apiClient := api.NewModularClient(accountID, token.AccessToken)
		projectOps := apiClient.Projects()

		// Fetch projects
		projects, err := projectOps.GetProjects(context.Background())
		if err != nil {
			return projectsLoadedMsg{err: err}
		}

		return projectsLoadedMsg{projects: projects}
	}
}

// accountItem implements list.Item
type accountItem struct {
	id   string
	name string
}

func (i accountItem) FilterValue() string { return i.name }
func (i accountItem) Title() string       { return i.name }
func (i accountItem) Description() string { return fmt.Sprintf("ID: %s", i.id) }

// projectItem implements list.Item
type projectItem struct {
	id   string
	name string
	desc string
}

func (i projectItem) FilterValue() string { return i.name }
func (i projectItem) Title() string       { return i.name }
func (i projectItem) Description() string { return i.desc }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Custom item delegate for cleaner rendering
type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(accountItem)
	if !ok {
		// Try project item
		if p, ok := listItem.(projectItem); ok {
			str := fmt.Sprintf("  %s", p.name)
			if index == m.Index() {
				_, _ = fmt.Fprint(w, selectedItemStyle.Render("→ "+p.name))
			} else {
				_, _ = fmt.Fprint(w, normalItemStyle.Render(str))
			}
			return
		}
		return
	}

	str := fmt.Sprintf("  %s", i.name)
	if index == m.Index() {
		_, _ = fmt.Fprint(w, selectedItemStyle.Render("→ "+i.name))
	} else {
		_, _ = fmt.Fprint(w, normalItemStyle.Render(str))
	}
}
