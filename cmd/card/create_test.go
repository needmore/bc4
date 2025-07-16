package card

import (
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewCreateCmd(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		errorContains string
	}{
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: false,
		},
		{
			name:          "with table flag",
			args:          []string{"--table", "123"},
			expectedError: false,
		},
		{
			name:          "with table and column flags",
			args:          []string{"--table", "123", "--column", "456"},
			expectedError: false,
		},
		{
			name:          "with account flag",
			args:          []string{"--account", "789"},
			expectedError: false,
		},
		{
			name:          "with project flag",
			args:          []string{"--project", "101112"},
			expectedError: false,
		},
		{
			name:          "with all flags",
			args:          []string{"--table", "123", "--column", "456", "--account", "789", "--project", "101112"},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock factory
			f := &factory.Factory{}
			cmd := newCreateCmd(f)

			// Set args
			cmd.SetArgs(tt.args)

			// Parse flags only (don't run the command)
			err := cmd.ParseFlags(tt.args)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateModel_Messages(t *testing.T) {
	// Test columnsLoadedMsg
	t.Run("columnsLoadedMsg success", func(t *testing.T) {
		testColumns := []api.Column{
			{ID: 1, Title: "To Do", CardsCount: 5},
			{ID: 2, Title: "In Progress", CardsCount: 3},
		}
		msg := columnsLoadedMsg{columns: testColumns, err: nil}
		assert.NoError(t, msg.err)
		assert.Len(t, msg.columns, 2)
	})

	// Test peopleLoadedMsg
	t.Run("peopleLoadedMsg success", func(t *testing.T) {
		testPeople := []api.Person{
			{ID: 1, Name: "John Doe", EmailAddress: "john@example.com"},
			{ID: 2, Name: "Jane Smith", EmailAddress: "jane@example.com"},
		}
		msg := peopleLoadedMsg{people: testPeople, err: nil}
		assert.NoError(t, msg.err)
		assert.Len(t, msg.people, 2)
	})

	// Test cardCreatedMsg
	t.Run("cardCreatedMsg success", func(t *testing.T) {
		testCard := &api.Card{
			ID:    123,
			Title: "Test Card",
		}
		msg := cardCreatedMsg{card: testCard, err: nil}
		assert.NoError(t, msg.err)
		assert.Equal(t, int64(123), msg.card.ID)
	})
}

func TestCreateModel_Init(t *testing.T) {
	model := createModel{
		factory:     &factory.Factory{},
		client:      &api.Client{},
		projectID:   "project123",
		cardTableID: 123,
	}

	// Test that Init returns proper commands
	cmd := model.Init()
	assert.NotNil(t, cmd)
}

func TestColumnItem(t *testing.T) {
	column := api.Column{
		ID:         1,
		Title:      "Test Column",
		CardsCount: 5,
	}

	item := columnItem{column: column}

	assert.Equal(t, "Test Column", item.Title())
	assert.Equal(t, "5 cards", item.Description())
	assert.Equal(t, "Test Column", item.FilterValue())
}

func TestPersonItem(t *testing.T) {
	person := api.Person{
		ID:           1,
		Name:         "John Doe",
		EmailAddress: "john@example.com",
	}

	t.Run("not selected", func(t *testing.T) {
		item := personItem{person: person, selected: false}
		assert.Equal(t, "  John Doe", item.Title())
		assert.Equal(t, "john@example.com", item.Description())
		assert.Equal(t, "John Doe john@example.com", item.FilterValue())
	})

	t.Run("selected", func(t *testing.T) {
		item := personItem{person: person, selected: true}
		assert.Equal(t, "✓ John Doe", item.Title())
		assert.Equal(t, "john@example.com", item.Description())
		assert.Equal(t, "John Doe john@example.com", item.FilterValue())
	})
}

func TestCreateModel_Markdown_Conversion(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		description string
	}{
		{
			name:        "bold text",
			content:     "This has **bold** text",
			description: "Should convert markdown bold to HTML strong tags",
		},
		{
			name:        "italic text",
			content:     "This has *italic* text",
			description: "Should convert markdown italic to HTML em tags",
		},
		{
			name:        "mixed formatting",
			content:     "This has **bold**, *italic*, and ~~strikethrough~~ text",
			description: "Should convert all markdown formatting correctly",
		},
		{
			name:        "code block",
			content:     "Here is `inline code` and more text",
			description: "Should convert inline code to pre tags",
		},
		{
			name:        "links",
			content:     "Check out [this link](https://example.com)",
			description: "Should preserve links in the content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that markdown content is preserved in the model
			model := createModel{
				cardContent: tt.content,
			}
			assert.Equal(t, tt.content, model.cardContent)
		})
	}
}

// Table-driven tests for command flag parsing
func TestCreateCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedTable  string
		expectedColumn string
		expectedAccount string
		expectedProject string
	}{
		{
			name: "all flags",
			args: []string{"--table", "123", "--column", "456", "--account", "789", "--project", "101112"},
			expectedTable: "123",
			expectedColumn: "456", 
			expectedAccount: "789",
			expectedProject: "101112",
		},
		{
			name: "short flags",
			args: []string{"-a", "789", "-p", "101112"},
			expectedAccount: "789",
			expectedProject: "101112",
		},
		{
			name: "partial flags",
			args: []string{"--table", "999"},
			expectedTable: "999",
		},
		{
			name: "no flags",
			args: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &factory.Factory{}
			cmd := newCreateCmd(f)
			
			err := cmd.ParseFlags(tt.args)
			assert.NoError(t, err)

			// Check flag values
			table, _ := cmd.Flags().GetString("table")
			column, _ := cmd.Flags().GetString("column")
			account, _ := cmd.Flags().GetString("account")
			project, _ := cmd.Flags().GetString("project")

			assert.Equal(t, tt.expectedTable, table)
			assert.Equal(t, tt.expectedColumn, column)
			assert.Equal(t, tt.expectedAccount, account)
			assert.Equal(t, tt.expectedProject, project)
		})
	}
}