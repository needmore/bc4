package card

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// handleAssigneeSelection handles the common logic for selecting/deselecting assignees
func handleAssigneeSelection(peopleList list.Model, selectedAssignees []int64, msg tea.Msg) (list.Model, []int64, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "space":
			if item, ok := peopleList.SelectedItem().(personItem); ok {
				// Toggle selection
				found := false
				for idx, id := range selectedAssignees {
					if id == item.person.ID {
						// Remove from selection
						selectedAssignees = append(selectedAssignees[:idx], selectedAssignees[idx+1:]...)
						found = true
						break
					}
				}
				if !found {
					// Add to selection
					selectedAssignees = append(selectedAssignees, item.person.ID)
				}
				// Update the list to reflect selection state
				items := peopleList.Items()
				for i, listItem := range items {
					if pi, ok := listItem.(personItem); ok {
						if pi.person.ID == item.person.ID {
							pi.selected = !found
							items[i] = pi
							break
						}
					}
				}
				peopleList.SetItems(items)
			}
		}
	}
	
	peopleList, cmd = peopleList.Update(msg)
	return peopleList, selectedAssignees, cmd
}