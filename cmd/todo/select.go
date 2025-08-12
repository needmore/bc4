package todo

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
)

type todoListsLoadedMsg struct {
	todoLists []api.TodoList
	err       error
}

type selectModel struct {
	list      list.Model
	todoLists []api.TodoList
	spinner   spinner.Model
	loading   bool
	err       error
	width     int
	height    int
	projectID string
	factory   *factory.Factory
}

func (m selectModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadTodoLists(),
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
			if selected, ok := m.list.SelectedItem().(todoListItem); ok {
				for _, tl := range m.todoLists {
					if strconv.FormatInt(tl.ID, 10) == selected.id {
						return m, tea.Sequence(
							m.saveDefaultTodoList(tl),
							tea.Quit,
						)
					}
				}
			}
		}

	case todoListsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.loading = false
			return m, tea.Quit
		}

		m.todoLists = msg.todoLists
		m.loading = false

		// Sort todo lists alphabetically
		sort.Slice(m.todoLists, func(i, j int) bool {
			return strings.ToLower(m.todoLists[i].Title) < strings.ToLower(m.todoLists[j].Title)
		})

		// Create list items
		items := make([]list.Item, 0, len(m.todoLists))
		for _, todoList := range m.todoLists {
			items = append(items, todoListItem{
				id:        strconv.FormatInt(todoList.ID, 10),
				name:      todoList.Title,
				desc:      todoList.Description,
				completed: todoList.CompletedRatio,
			})
		}

		// Calculate list dimensions
		listHeight := min(m.height-10, len(items)*3+6)
		listWidth := min(m.width-20, 70)

		// Create list with custom delegate
		m.list = list.New(items, todoListDelegate{}, listWidth, listHeight)
		m.list.Title = "Select Default Todo List"
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
		return fmt.Sprintf("\n  %s Loading todo lists...\n\n", m.spinner.View())
	}

	if len(m.todoLists) == 0 {
		return "\n  No todo lists found in this project.\n\n"
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

func (m *selectModel) loadTodoLists() tea.Cmd {
	return func() tea.Msg {
		// Create API client through factory
		apiClient, err := m.factory.ApiClient()
		if err != nil {
			return todoListsLoadedMsg{err: err}
		}
		todoOps := apiClient.Todos()

		// Get todo set for the project
		todoSet, err := todoOps.GetProjectTodoSet(m.factory.Context(), m.projectID)
		if err != nil {
			return todoListsLoadedMsg{err: fmt.Errorf("failed to get project todo set: %w", err)}
		}

		// Fetch todo lists
		todoLists, err := todoOps.GetTodoLists(m.factory.Context(), m.projectID, todoSet.ID)
		if err != nil {
			return todoListsLoadedMsg{err: err}
		}

		return todoListsLoadedMsg{todoLists: todoLists}
	}
}

func (m *selectModel) saveDefaultTodoList(todoList api.TodoList) tea.Cmd {
	return func() tea.Msg {
		cfg, err := m.factory.Config()
		if err != nil {
			return nil
		}

		// Get resolved account ID
		resolvedAccountID, err := m.factory.AccountID()
		if err != nil {
			return nil
		}

		// Update config
		if cfg.Accounts == nil {
			cfg.Accounts = make(map[string]config.AccountConfig)
		}

		acc := cfg.Accounts[resolvedAccountID]
		if acc.ProjectDefaults == nil {
			acc.ProjectDefaults = make(map[string]config.ProjectDefaults)
		}

		projDefaults := acc.ProjectDefaults[m.projectID]
		projDefaults.DefaultTodoList = strconv.FormatInt(todoList.ID, 10)
		acc.ProjectDefaults[m.projectID] = projDefaults
		cfg.Accounts[resolvedAccountID] = acc

		_ = config.Save(cfg)

		fmt.Printf("\nDefault todo list set to: %s (ID: %d)\n", todoList.Title, todoList.ID)
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

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))
)

// todoListItem implements list.Item
type todoListItem struct {
	id        string
	name      string
	desc      string
	completed string
}

func (i todoListItem) FilterValue() string { return i.name }
func (i todoListItem) Title() string       { return i.name }
func (i todoListItem) Description() string { return i.desc }

// Custom item delegate for cleaner rendering
type todoListDelegate struct{}

func (d todoListDelegate) Height() int                               { return 2 }
func (d todoListDelegate) Spacing() int                              { return 1 }
func (d todoListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d todoListDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(todoListItem)
	if !ok {
		return
	}

	// Render name with completion ratio
	name := i.name
	if i.completed != "" {
		name = fmt.Sprintf("%s %s", name, completedStyle.Render("("+i.completed+")"))
	}

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
	} else if i.completed != "" {
		// If no description, show last updated time if available
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(4)
		_, _ = fmt.Fprintln(w, descStyle.Render("Todo list"))
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
	var projectID string

	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select default todo list",
		Long:  `Interactively select a default todo list for the current project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply account override if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Apply project override if specified
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Create spinner
			s := spinner.New()
			s.Spinner = spinner.Dot
			s.Style = ui.SelectedItemStyle

			// Create model
			m := selectModel{
				spinner:   s,
				loading:   true,
				projectID: resolvedProjectID,
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
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID (overrides default)")

	return cmd
}
