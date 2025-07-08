package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

type step int

const (
	stepWelcome step = iota
	stepClientID
	stepClientSecret
	stepAuthenticate
	stepSelectAccount
	stepComplete
)

type authenticateMsg struct {
	token *auth.AccountToken
	err   error
}

type accountsLoadedMsg struct {
	accounts map[string]auth.AccountToken
}

// FirstRunModel represents the first-run wizard
type FirstRunModel struct {
	currentStep  step
	clientID     textinput.Model
	clientSecret textinput.Model
	authClient   *auth.Client
	token        *auth.AccountToken
	accounts     map[string]auth.AccountToken
	accountList  list.Model
	spinner      spinner.Model
	err          error
	width        int
	height       int
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
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyTab, tea.KeyShiftTab:
			if m.currentStep == stepClientID {
				m.clientID.Blur()
				m.clientSecret.Focus()
				m.currentStep = stepClientSecret
			} else if m.currentStep == stepClientSecret {
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
		m.currentStep = stepSelectAccount
		return m, m.loadAccounts()

	case accountsLoadedMsg:
		m.accounts = msg.accounts
		if len(m.accounts) > 1 {
			// Create account list
			items := make([]list.Item, 0, len(m.accounts))
			for _, account := range m.accounts {
				items = append(items, accountItem{
					id:   account.AccountID,
					name: account.AccountName,
				})
			}
			m.accountList = list.New(items, list.NewDefaultDelegate(), 0, 0)
			m.accountList.Title = "Select Default Account"
			m.accountList.SetShowStatusBar(false)
			m.accountList.SetFilteringEnabled(false)
			return m, nil
		}
		// Only one account, skip selection
		m.currentStep = stepComplete
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
		var cmd tea.Cmd
		m.accountList, cmd = m.accountList.Update(msg)
		return m, cmd
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
		titleStyle.Render("Authenticating..."),
		"",
		m.spinner.View()+" Opening browser for authentication",
		"",
		subtitleStyle.Render("Please authorize bc4 in your browser"),
		"",
		"If the browser doesn't open automatically,",
		"visit the URL shown in the terminal.",
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m FirstRunModel) viewSelectAccount() string {
	if len(m.accounts) <= 1 {
		return m.viewComplete()
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		m.accountList.View(),
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
			m.authClient.SetDefaultAccount(selected.id)
			m.currentStep = stepComplete
		}
		return m, nil

	case stepComplete:
		// Save credentials to config
		cfg := &config.Config{
			ClientID:     m.clientID.Value(),
			ClientSecret: m.clientSecret.Value(),
		}
		config.Save(cfg)
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

// accountItem implements list.Item
type accountItem struct {
	id   string
	name string
}

func (i accountItem) FilterValue() string { return i.name }
func (i accountItem) Title() string       { return i.name }
func (i accountItem) Description() string { return fmt.Sprintf("ID: %s", i.id) }