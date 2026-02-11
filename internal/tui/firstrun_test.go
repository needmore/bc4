package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestFirstRunModel_Update(t *testing.T) {
	t.Run("window resize", func(t *testing.T) {
		model := NewFirstRunModel()
		msg := tea.WindowSizeMsg{Width: 100, Height: 30}

		newModel, _ := model.Update(msg)
		m := newModel.(FirstRunModel)

		assert.Equal(t, 100, m.width)
		assert.Equal(t, 30, m.height)
	})

	t.Run("ctrl+c quits", func(t *testing.T) {
		model := NewFirstRunModel()
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}

		_, cmd := model.Update(msg)

		assert.NotNil(t, cmd)
		cmdMsg := cmd()
		_, isQuit := cmdMsg.(tea.QuitMsg)
		assert.True(t, isQuit)
	})

	t.Run("escape quits from welcome", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepWelcome
		msg := tea.KeyMsg{Type: tea.KeyEsc}

		_, cmd := model.Update(msg)

		assert.NotNil(t, cmd)
		cmdMsg := cmd()
		_, isQuit := cmdMsg.(tea.QuitMsg)
		assert.True(t, isQuit)
	})

	t.Run("enter advances from welcome to clientID", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepWelcome
		msg := tea.KeyMsg{Type: tea.KeyEnter}

		newModel, _ := model.Update(msg)
		m := newModel.(FirstRunModel)

		assert.Equal(t, stepClientID, m.currentStep)
	})

	t.Run("spinner tick updates spinner", func(t *testing.T) {
		model := NewFirstRunModel()
		msg := spinner.TickMsg{Time: time.Now()}

		_, cmd := model.Update(msg)

		assert.NotNil(t, cmd)
	})
}

func TestFirstRunModel_UninitializedListNoPanic(t *testing.T) {
	t.Run("stepSelectProject with uninitialized list does not panic on key msg", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectProject
		// projectList is zero-value (not initialized) - this is the bug scenario

		msg := tea.KeyMsg{Type: tea.KeyDown}

		assert.NotPanics(t, func() {
			model.Update(msg)
		})
	})

	t.Run("stepSelectProject with uninitialized list does not panic on window resize", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectProject

		msg := tea.WindowSizeMsg{Width: 80, Height: 24}

		assert.NotPanics(t, func() {
			model.Update(msg)
		})
	})

	t.Run("stepSelectProject with uninitialized list does not panic on spinner tick", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectProject

		msg := spinner.TickMsg{Time: time.Now()}

		assert.NotPanics(t, func() {
			model.Update(msg)
		})
	})

	t.Run("stepSelectAccount with uninitialized list does not panic on key msg", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectAccount
		// accountList is zero-value (not initialized)

		msg := tea.KeyMsg{Type: tea.KeyDown}

		assert.NotPanics(t, func() {
			model.Update(msg)
		})
	})

	t.Run("stepSelectAccount with uninitialized list does not panic on window resize", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectAccount

		msg := tea.WindowSizeMsg{Width: 80, Height: 24}

		assert.NotPanics(t, func() {
			model.Update(msg)
		})
	})

	t.Run("stepSelectAccount with uninitialized list does not panic on spinner tick", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectAccount

		msg := spinner.TickMsg{Time: time.Now()}

		assert.NotPanics(t, func() {
			model.Update(msg)
		})
	})
}

func TestFirstRunModel_InitializedListForwardsMessages(t *testing.T) {
	t.Run("stepSelectProject forwards messages when list is populated", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectProject

		items := []list.Item{
			projectItem{id: "1", name: "Project A", desc: "Desc A"},
			projectItem{id: "2", name: "Project B", desc: "Desc B"},
		}
		model.projectList = list.New(items, itemDelegate{}, 60, 10)

		msg := tea.KeyMsg{Type: tea.KeyDown}

		assert.NotPanics(t, func() {
			newModel, _ := model.Update(msg)
			m := newModel.(FirstRunModel)
			// List should have processed the message
			assert.Equal(t, stepSelectProject, m.currentStep)
			assert.Equal(t, 2, len(m.projectList.Items()))
		})
	})

	t.Run("stepSelectAccount forwards messages when list is populated", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectAccount

		items := []list.Item{
			accountItem{id: "1", name: "Account A"},
			accountItem{id: "2", name: "Account B"},
		}
		model.accountList = list.New(items, itemDelegate{}, 50, 10)

		msg := tea.KeyMsg{Type: tea.KeyDown}

		assert.NotPanics(t, func() {
			newModel, _ := model.Update(msg)
			m := newModel.(FirstRunModel)
			assert.Equal(t, stepSelectAccount, m.currentStep)
			assert.Equal(t, 2, len(m.accountList.Items()))
		})
	})
}

func TestFirstRunModel_StepTransitions(t *testing.T) {
	t.Run("tab switches from clientID to clientSecret", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepClientID
		msg := tea.KeyMsg{Type: tea.KeyTab}

		newModel, _ := model.Update(msg)
		m := newModel.(FirstRunModel)

		assert.Equal(t, stepClientSecret, m.currentStep)
	})

	t.Run("shift-tab switches from clientSecret to clientID", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepClientSecret
		msg := tea.KeyMsg{Type: tea.KeyShiftTab}

		newModel, _ := model.Update(msg)
		m := newModel.(FirstRunModel)

		assert.Equal(t, stepClientID, m.currentStep)
	})

	t.Run("enter with empty clientID shows error", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepClientID
		msg := tea.KeyMsg{Type: tea.KeyEnter}

		newModel, _ := model.Update(msg)
		m := newModel.(FirstRunModel)

		assert.NotNil(t, m.err)
		assert.Equal(t, stepClientID, m.currentStep)
	})

	t.Run("enter with clientID advances to clientSecret", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepClientID
		model.clientID.SetValue("test-client-id")
		msg := tea.KeyMsg{Type: tea.KeyEnter}

		newModel, _ := model.Update(msg)
		m := newModel.(FirstRunModel)

		assert.Equal(t, stepClientSecret, m.currentStep)
	})

	t.Run("enter with empty clientSecret shows error", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepClientSecret
		msg := tea.KeyMsg{Type: tea.KeyEnter}

		newModel, _ := model.Update(msg)
		m := newModel.(FirstRunModel)

		assert.NotNil(t, m.err)
		assert.Equal(t, stepClientSecret, m.currentStep)
	})

	t.Run("enter on complete step quits", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepComplete
		msg := tea.KeyMsg{Type: tea.KeyEnter}

		_, cmd := model.Update(msg)

		assert.NotNil(t, cmd)
		cmdMsg := cmd()
		_, isQuit := cmdMsg.(tea.QuitMsg)
		assert.True(t, isQuit)
	})
}

func TestFirstRunModel_View(t *testing.T) {
	t.Run("welcome view", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepWelcome
		model.width = 80
		model.height = 24

		view := model.View()
		assert.Contains(t, view, "Welcome to bc4!")
		assert.Contains(t, view, "Press Enter to continue")
	})

	t.Run("clientID view", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepClientID
		model.width = 80
		model.height = 24

		view := model.View()
		assert.Contains(t, view, "OAuth Setup")
		assert.Contains(t, view, "Client ID")
	})

	t.Run("clientSecret view", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepClientSecret
		model.width = 80
		model.height = 24

		view := model.View()
		assert.Contains(t, view, "Client Secret")
	})

	t.Run("authenticate view", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepAuthenticate
		model.width = 80
		model.height = 24

		view := model.View()
		assert.Contains(t, view, "Authenticating")
		assert.Contains(t, view, "browser")
	})

	t.Run("select account loading view", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectAccount
		model.width = 80
		model.height = 24

		view := model.View()
		assert.Contains(t, view, "Loading accounts")
	})

	t.Run("select project loading view", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepSelectProject
		model.width = 80
		model.height = 24

		view := model.View()
		assert.Contains(t, view, "Loading projects")
	})

	t.Run("complete view", func(t *testing.T) {
		model := NewFirstRunModel()
		model.currentStep = stepComplete
		model.width = 80
		model.height = 24

		view := model.View()
		assert.Contains(t, view, "Setup Complete")
	})
}
