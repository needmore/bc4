package card

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/attachments"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/mentions"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

type editStep int

const (
	editStepLoading editStep = iota
	editStepEditTitle
	editStepEditContent
	editStepSelectAssignees
	editStepConfirm
	editStepUpdating
	editStepDone
)

type cardLoadedMsg struct {
	card *api.Card
	err  error
}

type cardUpdatedMsg struct {
	card *api.Card
	err  error
}

type editModel struct {
	factory           *factory.Factory
	client            *api.Client
	projectID         string
	cardID            int64
	step              editStep
	card              *api.Card
	people            []api.Person
	peopleList        list.Model
	titleInput        textinput.Model
	contentArea       textarea.Model
	spinner           spinner.Model
	selectedAssignees []int64
	cardTitle         string
	cardContent       string
	updatedCard       *api.Card
	err               error
	width             int
	height            int
}

func (m editModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadCard(),
		m.loadPeople(),
	)
}

func (m editModel) loadCard() tea.Cmd {
	return func() tea.Msg {
		card, err := m.client.GetCard(m.factory.Context(), m.projectID, m.cardID)
		if err != nil {
			return cardLoadedMsg{err: err}
		}
		return cardLoadedMsg{card: card}
	}
}

func (m editModel) loadPeople() tea.Cmd {
	return func() tea.Msg {
		people, err := m.client.GetProjectPeople(m.factory.Context(), m.projectID)
		if err != nil {
			return peopleLoadedMsg{err: err}
		}
		return peopleLoadedMsg{people: people}
	}
}

func (m editModel) updateCard() tea.Cmd {
	return func() tea.Msg {
		// Convert content to rich text
		converter := markdown.NewConverter()
		richContent, err := converter.MarkdownToRichText(m.cardContent)
		if err != nil {
			return cardUpdatedMsg{err: err}
		}

		// Replace inline @Name mentions with bc-attachment tags
		richContent, err = mentions.Resolve(m.factory.Context(), richContent, m.client, m.projectID)
		if err != nil {
			return cardUpdatedMsg{err: err}
		}

		req := api.CardUpdateRequest{
			Title:       m.cardTitle,
			Content:     richContent,
			AssigneeIDs: m.selectedAssignees,
		}

		card, err := m.client.UpdateCard(m.factory.Context(), m.projectID, m.cardID, req)
		if err != nil {
			return cardUpdatedMsg{err: err}
		}
		return cardUpdatedMsg{card: card}
	}
}

func (m editModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.peopleList.SetSize(msg.Width, msg.Height-10)
		m.contentArea.SetWidth(msg.Width - 4)
		m.contentArea.SetHeight(msg.Height - 10)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.step != editStepEditContent {
				return m, tea.Quit
			}
		}

	case cardLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.card = msg.card
		m.cardTitle = m.card.Title
		m.cardContent = m.card.Content // This is HTML, ideally we'd convert to markdown

		// Set initial values
		m.titleInput.SetValue(m.cardTitle)
		m.contentArea.SetValue(m.cardContent)

		// Extract current assignees
		for _, assignee := range m.card.Assignees {
			m.selectedAssignees = append(m.selectedAssignees, assignee.ID)
		}

		m.step = editStepEditTitle
		m.titleInput.Focus()
		cmds = append(cmds, textinput.Blink)

	case peopleLoadedMsg:
		if msg.err == nil {
			m.people = msg.people
		}

	case cardUpdatedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.updatedCard = msg.card
		m.step = editStepDone
		return m, tea.Quit

	case spinner.TickMsg:
		if m.step == editStepLoading || m.step == editStepUpdating {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Handle step-specific updates
	switch m.step {
	case editStepEditTitle:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.cardTitle = m.titleInput.Value()
				m.step = editStepEditContent
				m.contentArea.Focus()
			}
		}
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)

	case editStepEditContent:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+d":
				m.cardContent = m.contentArea.Value()
				if len(m.people) > 0 {
					// Setup people list
					items := make([]list.Item, len(m.people))
					for i, person := range m.people {
						selected := false
						for _, id := range m.selectedAssignees {
							if id == person.ID {
								selected = true
								break
							}
						}
						items[i] = personItem{person: person, selected: selected}
					}
					m.peopleList.SetItems(items)
					m.step = editStepSelectAssignees
				} else {
					m.step = editStepConfirm
				}
			}
		}
		m.contentArea, cmd = m.contentArea.Update(msg)
		cmds = append(cmds, cmd)

	case editStepSelectAssignees:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				m.step = editStepConfirm
			}
		}
		m.peopleList, m.selectedAssignees, cmd = handleAssigneeSelection(m.peopleList, m.selectedAssignees, msg)
		cmds = append(cmds, cmd)

	case editStepConfirm:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				m.step = editStepUpdating
				cmds = append(cmds, m.updateCard())
			case "n", "N":
				return m, tea.Quit
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m editModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.step == editStepDone && m.updatedCard != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(fmt.Sprintf("Card updated: #%d", m.updatedCard.ID))
	}

	var content string
	title := lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1)

	switch m.step {
	case editStepLoading:
		content = m.spinner.View() + " Loading card..."

	case editStepEditTitle:
		content = title.Render("Edit Card Title") + "\n\n"
		content += m.titleInput.View()
		content += "\n\nPress Enter to continue"

	case editStepEditContent:
		content = title.Render("Edit Card Content") + "\n\n"
		content += "Note: Content is shown as HTML. Markdown support coming soon.\n\n"
		content += m.contentArea.View()
		content += "\n\nPress Ctrl+D when done, Esc to keep current content"

	case editStepSelectAssignees:
		content = title.Render("Edit Assignees") + "\n\n"
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

	case editStepConfirm:
		content = title.Render("Confirm Changes") + "\n\n"
		if m.card != nil {
			if m.cardTitle != m.card.Title {
				content += fmt.Sprintf("Title: %s → %s\n", m.card.Title, m.cardTitle)
			} else {
				content += fmt.Sprintf("Title: %s (unchanged)\n", m.cardTitle)
			}

			if m.cardContent != m.card.Content {
				content += "Content: Modified\n"
			} else {
				content += "Content: Unchanged\n"
			}

			// Compare assignees
			oldAssigneeIDs := make(map[int64]bool)
			for _, a := range m.card.Assignees {
				oldAssigneeIDs[a.ID] = true
			}

			newAssigneeIDs := make(map[int64]bool)
			for _, id := range m.selectedAssignees {
				newAssigneeIDs[id] = true
			}

			assigneesChanged := len(oldAssigneeIDs) != len(newAssigneeIDs)
			if !assigneesChanged {
				for id := range oldAssigneeIDs {
					if !newAssigneeIDs[id] {
						assigneesChanged = true
						break
					}
				}
			}

			if assigneesChanged {
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
			} else {
				content += "Assignees: Unchanged\n"
			}
		}
		content += "\nUpdate card? (y/n)"

	case editStepUpdating:
		content = m.spinner.View() + " Updating card..."
	}

	return content
}

func updateCardNonInteractive(f *factory.Factory, client *api.Client, projectID string, cardID int64, title, content string, attach []string) error {
	// Get current card
	card, err := client.GetCard(f.Context(), projectID, cardID)
	if err != nil {
		return fmt.Errorf("failed to get card: %w", err)
	}

	// Build update request
	req := api.CardUpdateRequest{}

	if title != "" {
		req.Title = title
	} else {
		req.Title = card.Title
	}

	if content != "" {
		// Convert markdown to rich text
		converter := markdown.NewConverter()
		richContent, err := converter.MarkdownToRichText(content)
		if err != nil {
			return fmt.Errorf("failed to convert content: %w", err)
		}

		// Replace inline @Name mentions with bc-attachment tags
		richContent, err = mentions.Resolve(f.Context(), richContent, client, projectID)
		if err != nil {
			return fmt.Errorf("failed to resolve mentions: %w", err)
		}

		req.Content = richContent
	} else {
		req.Content = card.Content
	}

	// Handle attachments - append to existing or new content
	if len(attach) > 0 {
		for _, attachPath := range attach {
			fileData, err := os.ReadFile(attachPath)
			if err != nil {
				return fmt.Errorf("failed to read attachment %s: %w", attachPath, err)
			}
			filename := filepath.Base(attachPath)
			upload, err := client.UploadAttachment(filename, fileData, "")
			if err != nil {
				return fmt.Errorf("failed to upload attachment %s: %w", filename, err)
			}
			tag := attachments.BuildTag(upload.AttachableSGID)
			req.Content += tag
		}
	}

	// Preserve assignees
	assigneeIDs := make([]int64, 0, len(card.Assignees))
	for _, a := range card.Assignees {
		assigneeIDs = append(assigneeIDs, a.ID)
	}
	req.AssigneeIDs = assigneeIDs

	// Update the card
	updatedCard, err := client.UpdateCard(f.Context(), projectID, cardID, req)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	fmt.Printf("Card #%d updated\n", updatedCard.ID)
	return nil
}

func newEditCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var attach []string

	cmd := &cobra.Command{
		Use:   "edit [ID or URL]",
		Short: "Edit card title/content",
		Long: `Edit the title and content of an existing card.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Use --attach to add images or files to the card content. Attachments are
appended to the existing content. Multiple files can be attached by using
the flag multiple times.`,
		Example: `  # Edit card title
  bc4 card edit 12345 --title "New title"

  # Edit card content
  bc4 card edit 12345 --content "Updated description"

  # Add an image attachment to an existing card
  bc4 card edit 12345 --attach ./screenshot.png

  # Add multiple attachments
  bc4 card edit 12345 --attach ./photo1.jpg --attach ./photo2.jpg`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// If a URL was parsed, override account and project IDs if provided
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeCard {
					return fmt.Errorf("URL is not for a card: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
				}
				if parsedURL.ProjectID > 0 {
					f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
				}
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Check for non-interactive flags
			title, _ := cmd.Flags().GetString("title")
			content, _ := cmd.Flags().GetString("content")
			interactive, _ := cmd.Flags().GetBool("interactive")

			// If non-interactive mode with flags
			if !interactive && (title != "" || content != "" || len(attach) > 0) {
				return updateCardNonInteractive(f, client.Client, resolvedProjectID, cardID, title, content, attach)
			}

			// Interactive mode
			model := editModel{
				factory:     f,
				client:      client.Client,
				projectID:   resolvedProjectID,
				cardID:      cardID,
				step:        editStepLoading,
				spinner:     spinner.New(),
				titleInput:  textinput.New(),
				contentArea: textarea.New(),
			}

			// Configure inputs
			model.titleInput.Placeholder = "Enter card title..."
			model.titleInput.CharLimit = 200
			model.contentArea.Placeholder = "Enter card content (Markdown supported)..."
			model.contentArea.CharLimit = 10000

			// Configure people list
			model.peopleList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
			model.peopleList.Title = "Select Assignees"
			model.peopleList.SetShowStatusBar(false)
			model.peopleList.SetFilteringEnabled(true)

			// Run the program
			p := tea.NewProgram(model, tea.WithAltScreen())
			finalModel, err := p.Run()
			if err != nil {
				return err
			}

			// Check if card was updated
			if m, ok := finalModel.(editModel); ok {
				if m.err != nil {
					return m.err
				}
				if m.updatedCard != nil {
					fmt.Printf("Card #%d updated\n", m.updatedCard.ID)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().String("title", "", "New title for the card")
	cmd.Flags().String("content", "", "New content for the card (Markdown supported)")
	cmd.Flags().Bool("interactive", false, "Use interactive mode (default when no flags)")
	cmd.Flags().StringSliceVar(&attach, "attach", nil, "Attach file(s) to the card (can be used multiple times)")

	return cmd
}
