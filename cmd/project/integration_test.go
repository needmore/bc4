package project

import (
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

// TestProjectCommands performs integration tests on project commands
// These tests verify the commands are properly constructed but don't execute them
// due to the complexity of mocking the factory dependencies
func TestProjectCommands(t *testing.T) {
	f := factory.New()

	t.Run("list command", func(t *testing.T) {
		cmd := newListCmd(f)
		assert.Equal(t, "list", cmd.Use)
		assert.Contains(t, cmd.Aliases, "ls")
		assert.NotNil(t, cmd.RunE)

		// Test flags
		assert.NotNil(t, cmd.Flag("json"))
		assert.NotNil(t, cmd.Flag("format"))
	})

	t.Run("search command", func(t *testing.T) {
		cmd := newSearchCmd(f)
		assert.Equal(t, "search [query]", cmd.Use)
		assert.NotNil(t, cmd.RunE)
		// Can't compare function values directly

		// Test flags
		assert.NotNil(t, cmd.Flag("json"))
		assert.NotNil(t, cmd.Flag("account"))
	})

	t.Run("set command", func(t *testing.T) {
		cmd := newSetCmd(f)
		assert.Equal(t, "set [project-id]", cmd.Use)
		assert.NotNil(t, cmd.RunE)
		// Can't compare function values directly

		// Test flags
		// No clear flag in set command
	})

	t.Run("view command", func(t *testing.T) {
		cmd := newViewCmd(f)
		assert.Equal(t, "view [project-id or URL]", cmd.Use)
		assert.NotNil(t, cmd.RunE)
		// Can't compare function values directly

		// Test flags
		assert.NotNil(t, cmd.Flag("json"))
	})

	t.Run("select command", func(t *testing.T) {
		cmd := newSelectCmd(f)
		assert.Equal(t, "select", cmd.Use)
		assert.NotNil(t, cmd.RunE)
		// Can't compare function values directly
	})
}

// TestSortProjectsByName tests the sorting function
func TestSortProjectsByNameFunc(t *testing.T) {
	// This is already tested in list_test.go
	// Just verify the function exists
	projects := []api.Project{
		{ID: 2, Name: "Beta"},
		{ID: 1, Name: "Alpha"},
	}

	sortProjectsByName(projects)
	assert.Equal(t, "Alpha", projects[0].Name)
	assert.Equal(t, "Beta", projects[1].Name)
}
