package card

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/spf13/cobra"
)

type createStep int

const (
	stepSelectColumn createStep = iota
	stepEnterTitle
	stepEnterContent
	stepSelectAssignees
	stepConfirm
	stepCreating
	stepDone
)

// columnItem wraps api.Column to implement list.Item interface
type columnItem struct {
	column api.Column
}

func (i columnItem) Title() string       { return i.column.Title }
func (i columnItem) Description() string { return fmt.Sprintf("%d cards", i.column.CardsCount) }
func (i columnItem) FilterValue() string { return i.column.Title }

// personItem wraps api.Person to implement list.Item interface
type personItem struct {
	person   api.Person
	selected bool
}

func (i personItem) Title() string {
	prefix := "  "
	if i.selected {
		prefix = "✓ "
	}
	return prefix + i.person.Name
}

func (i personItem) Description() string { return i.person.EmailAddress }
func (i personItem) FilterValue() string { return i.person.Name + " " + i.person.EmailAddress }

type cardCreatedMsg struct {
	card *api.Card
	err  error
}

type columnsLoadedMsg struct {
	columns []api.Column
	err     error
}

type peopleLoadedMsg struct {
	people []api.Person
	err    error
}

type createModel struct {
	factory           *factory.Factory
	client            *api.Client
	projectID         string
	cardTableID       int64
	step              createStep
	columns           []api.Column
	people            []api.Person
	columnList        list.Model
	peopleList        list.Model
	titleInput        textinput.Model
	contentInput      textinput.Model
	spinner           spinner.Model
	selectedColumn    *api.Column
	selectedAssignees []int64
	cardTitle         string
	cardContent       string
	createdCard       *api.Card
	err               error
	width             int
	height            int
}

func (m createModel) Init() tea.Cmd {
	// Load columns when starting
	return tea.Batch(
		m.spinner.Tick,
		m.loadColumns(),
		m.loadPeople(),
	)
}

func (m createModel) loadColumns() tea.Cmd {
	return func() tea.Msg {
		cardTable, err := m.client.GetCardTable(m.factory.Context(), m.projectID, m.cardTableID)
		if err != nil {
			return columnsLoadedMsg{err: err}
		}
		return columnsLoadedMsg{columns: cardTable.Lists}
	}
}

func (m createModel) loadPeople() tea.Cmd {
	return func() tea.Msg {
		people, err := m.client.GetProjectPeople(m.factory.Context(), m.projectID)
		if err != nil {
			return peopleLoadedMsg{err: err}
		}
		return peopleLoadedMsg{people: people}
	}
}

func (m createModel) createCard() tea.Cmd {
	return func() tea.Msg {
		// Convert content to rich text
		converter := markdown.NewConverter()
		richContent, err := converter.MarkdownToRichText(m.cardContent)
		if err != nil {
			return cardCreatedMsg{err: err}
		}

		req := api.CardCreateRequest{
			Title:   m.cardTitle,
			Content: richContent,
		}

		card, err := m.client.CreateCard(m.factory.Context(), m.projectID, m.selectedColumn.ID, req)
		if err != nil {
			return cardCreatedMsg{err: err}
		}

		// If assignees were selected, update the card to add them
		if len(m.selectedAssignees) > 0 {
			updateReq := api.CardUpdateRequest{
				Title:       m.cardTitle,
				Content:     richContent,
				AssigneeIDs: m.selectedAssignees,
			}
			card, err = m.client.UpdateCard(m.factory.Context(), m.projectID, card.ID, updateReq)
			if err != nil {
				// Card was created but assigning failed
				return cardCreatedMsg{card: card, err: fmt.Errorf("card created but failed to assign: %w", err)}
			}
		}

		return cardCreatedMsg{card: card}
	}
}

func (m createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.columnList.SetSize(msg.Width, msg.Height-10)
		m.peopleList.SetSize(msg.Width, msg.Height-10)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}

	case columnsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.columns = msg.columns
		// Setup column list
		items := make([]list.Item, len(m.columns))
		for i, col := range m.columns {
			items[i] = columnItem{column: col}
		}
		m.columnList.SetItems(items)

	case peopleLoadedMsg:
		if msg.err != nil {
			// People loading is not critical, continue without assignees
			m.people = []api.Person{}
		} else {
			m.people = msg.people
		}

	case cardCreatedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.createdCard = msg.card
		m.step = stepDone
		return m, tea.Quit

	case spinner.TickMsg:
		if m.step == stepCreating {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Handle step-specific updates
	switch m.step {
	case stepSelectColumn:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if i, ok := m.columnList.SelectedItem().(columnItem); ok {
					m.selectedColumn = &i.column
					m.step = stepEnterTitle
					m.titleInput.Focus()
					cmds = append(cmds, textinput.Blink)
				}
			}
		}
		m.columnList, cmd = m.columnList.Update(msg)
		cmds = append(cmds, cmd)

	case stepEnterTitle:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if m.titleInput.Value() != "" {
					m.cardTitle = m.titleInput.Value()
					m.step = stepEnterContent
					m.contentInput.Focus()
					cmds = append(cmds, textinput.Blink)
				}
			}
		}
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)

	case stepEnterContent:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+d":
				m.cardContent = m.contentInput.Value()
				if len(m.people) > 0 {
					// Setup people list
					items := make([]list.Item, len(m.people))
					for i, person := range m.people {
						items[i] = personItem{person: person, selected: false}
					}
					m.peopleList.SetItems(items)
					m.step = stepSelectAssignees
				} else {
					m.step = stepConfirm
				}
			}
		}
		m.contentInput, cmd = m.contentInput.Update(msg)
		cmds = append(cmds, cmd)

	case stepSelectAssignees:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "space":
				if item, ok := m.peopleList.SelectedItem().(personItem); ok {
					// Toggle selection
					found := false
					for idx, id := range m.selectedAssignees {
						if id == item.person.ID {
							// Remove from selection
							m.selectedAssignees = append(m.selectedAssignees[:idx], m.selectedAssignees[idx+1:]...)
							found = true
							break
						}
					}
					if !found {
						// Add to selection
						m.selectedAssignees = append(m.selectedAssignees, item.person.ID)
					}
					// Update the list to reflect selection state
					items := m.peopleList.Items()
					for i, listItem := range items {
						if pi, ok := listItem.(personItem); ok {
							if pi.person.ID == item.person.ID {
								pi.selected = !found
								items[i] = pi
								break
							}
						}
					}
					m.peopleList.SetItems(items)
				}
			case "enter":
				m.step = stepConfirm
			}
		}
		m.peopleList, cmd = m.peopleList.Update(msg)
		cmds = append(cmds, cmd)

	case stepConfirm:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				m.step = stepCreating
				cmds = append(cmds, m.createCard())
			case "n", "N":
				return m, tea.Quit
			}
		}

	case stepCreating:
		// Just update spinner
	}

	return m, tea.Batch(cmds...)
}

func (m createModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.step == stepDone && m.createdCard != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(fmt.Sprintf("Card created: #%d", m.createdCard.ID))
	}

	var content string
	title := lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1)

	switch m.step {
	case stepSelectColumn:
		content = title.Render("Select Column") + "\n\n"
		if len(m.columns) == 0 {
			content += m.spinner.View() + " Loading columns..."
		} else {
			content += m.columnList.View()
		}

	case stepEnterTitle:
		content = title.Render("Enter Card Title") + "\n\n"
		content += m.titleInput.View()
		content += "\n\nPress Enter to continue"

	case stepEnterContent:
		content = title.Render("Enter Card Content (Markdown supported)") + "\n\n"
		content += m.contentInput.View()
		content += "\n\nPress Ctrl+D when done"

	case stepSelectAssignees:
		content = title.Render("Select Assignees") + "\n\n"
		content += "Press Space to toggle selection, Enter when done\n\n"

		// Show selected assignees
		if len(m.selectedAssignees) > 0 {
			content += "Selected: "
			var names []string
			for _, id := range m.selectedAssignees {
				for _, p := range m.people {
					if p.ID == id {
						names = append(names, p.Name)
						break
					}
				}
			}
			content += strings.Join(names, ", ") + "\n\n"
		}

		content += m.peopleList.View()

	case stepConfirm:
		content = title.Render("Confirm Card Creation") + "\n\n"
		content += fmt.Sprintf("Column: %s\n", m.selectedColumn.Title)
		content += fmt.Sprintf("Title: %s\n", m.cardTitle)
		content += fmt.Sprintf("Content: %s\n", m.cardContent)
		if len(m.selectedAssignees) > 0 {
			var names []string
			for _, id := range m.selectedAssignees {
				for _, p := range m.people {
					if p.ID == id {
						names = append(names, p.Name)
						break
					}
				}
			}
			content += fmt.Sprintf("Assignees: %s\n", strings.Join(names, ", "))
		}
		content += "\nCreate card? (y/n)"

	case stepCreating:
		content = m.spinner.View() + " Creating card..."
	}

	return content
}

func newCreateCmd(f *factory.Factory) *cobra.Command {
	var cardTableID string
	var columnID string
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new card interactively",
		Long: `Create a new card using an interactive interface.

If you specify a card table ID, the interactive UI will start from column selection.
If you also specify a column ID, it will skip to entering card details.

Examples:
  bc4 card create                      # Full interactive mode
  bc4 card create --table 123          # Start from column selection in table 123  
  bc4 card create --table 123 --column 456  # Skip to card details for column 456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Get API client
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// If no table ID specified, we need to find the project's card table
			var tableID int64
			if cardTableID != "" {
				tableID, err = strconv.ParseInt(cardTableID, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid card table ID: %s", cardTableID)
				}
			} else {
				// Get the project's card table
				cardTable, err := client.Cards().GetProjectCardTable(f.Context(), resolvedProjectID)
				if err != nil {
					return fmt.Errorf("failed to get project card table: %w", err)
				}
				tableID = cardTable.ID
			}

			// Initialize the model
			model := createModel{
				factory:      f,
				client:       client.Client,
				projectID:    resolvedProjectID,
				cardTableID:  tableID,
				step:         stepSelectColumn,
				spinner:      spinner.New(),
				titleInput:   textinput.New(),
				contentInput: textinput.New(),
			}

			// Configure inputs
			model.titleInput.Placeholder = "Enter card title..."
			model.titleInput.CharLimit = 200
			model.contentInput.Placeholder = "Enter card content (Markdown supported)..."
			model.contentInput.CharLimit = 5000

			// Configure lists
			model.columnList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
			model.columnList.Title = "Select Column"
			model.columnList.SetShowStatusBar(false)
			model.columnList.SetFilteringEnabled(false)

			model.peopleList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
			model.peopleList.Title = "Select Assignees"
			model.peopleList.SetShowStatusBar(false)
			model.peopleList.SetFilteringEnabled(true)

			// If column ID is specified, skip column selection
			if columnID != "" {
				colID, err := strconv.ParseInt(columnID, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid column ID: %s", columnID)
				}
				// We need to load the column details
				// For simplicity, we'll still load all columns but pre-select this one
				model.selectedColumn = &api.Column{ID: colID}
				model.step = stepEnterTitle
				model.titleInput.Focus()
			}

			// Run the program
			p := tea.NewProgram(model, tea.WithAltScreen())
			finalModel, err := p.Run()
			if err != nil {
				return err
			}

			// Check if card was created
			if m, ok := finalModel.(createModel); ok {
				if m.err != nil {
					return m.err
				}
				if m.createdCard != nil {
					fmt.Printf("#%d\n", m.createdCard.ID)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&cardTableID, "table", "", "Card table ID")
	cmd.Flags().StringVar(&columnID, "column", "", "Column ID (requires --table)")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
