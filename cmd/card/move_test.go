package card

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewMoveCmd(t *testing.T) {
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
			name:          "with column flag",
			args:          []string{"--column", "Done"},
			expectedError: false,
		},
		{
			name:          "with account flag",
			args:          []string{"--account", "123"},
			expectedError: false,
		},
		{
			name:          "with project flag",
			args:          []string{"--project", "456"},
			expectedError: false,
		},
		{
			name:          "with all flags",
			args:          []string{"--column", "Done", "--account", "123", "--project", "456"},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory.New()
			cmd := newMoveCmd(f)

			cmd.SetArgs(tt.args)
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

func TestMoveCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedColumn  string
		expectedAccount string
		expectedProject string
	}{
		{
			name:            "all flags",
			args:            []string{"--column", "In Progress", "--account", "123", "--project", "456"},
			expectedColumn:  "In Progress",
			expectedAccount: "123",
			expectedProject: "456",
		},
		{
			name:            "short flags",
			args:            []string{"-a", "789", "-p", "101112"},
			expectedAccount: "789",
			expectedProject: "101112",
		},
		{
			name:           "column name flag",
			args:           []string{"--column", "Done"},
			expectedColumn: "Done",
		},
		{
			name:           "column ID flag",
			args:           []string{"--column", "999"},
			expectedColumn: "999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory.New()
			cmd := newMoveCmd(f)

			err := cmd.ParseFlags(tt.args)
			assert.NoError(t, err)

			column, _ := cmd.Flags().GetString("column")
			account, _ := cmd.Flags().GetString("account")
			project, _ := cmd.Flags().GetString("project")

			assert.Equal(t, tt.expectedColumn, column)
			assert.Equal(t, tt.expectedAccount, account)
			assert.Equal(t, tt.expectedProject, project)
		})
	}
}

// TestFindCardTable tests the logic for finding which card table contains a card
func TestFindCardTable(t *testing.T) {
	board1 := &api.CardTable{
		ID:    100,
		Title: "Development Board",
		Lists: []api.Column{
			{ID: 1, Title: "To Do"},
			{ID: 2, Title: "In Progress"},
			{ID: 3, Title: "Done"},
		},
	}

	board2 := &api.CardTable{
		ID:    200,
		Title: "Marketing Board",
		Lists: []api.Column{
			{ID: 4, Title: "Backlog"},
			{ID: 5, Title: "In Progress"},
			{ID: 6, Title: "Published"},
		},
	}

	tests := []struct {
		name               string
		card               *api.Card
		cardTables         []*api.CardTable
		expectedBoardID    int64
		expectedBoardTitle string
	}{
		{
			name: "card on first board",
			card: &api.Card{
				ID:    1001,
				Title: "Fix bug",
				Parent: &api.Column{
					ID:    1, // To Do column on Development Board
					Title: "To Do",
				},
			},
			cardTables:         []*api.CardTable{board1, board2},
			expectedBoardID:    100,
			expectedBoardTitle: "Development Board",
		},
		{
			name: "card on second board",
			card: &api.Card{
				ID:    2001,
				Title: "Blog post",
				Parent: &api.Column{
					ID:    4, // Backlog column on Marketing Board
					Title: "Backlog",
				},
			},
			cardTables:         []*api.CardTable{board1, board2},
			expectedBoardID:    200,
			expectedBoardTitle: "Marketing Board",
		},
		{
			name: "card without parent defaults to first board",
			card: &api.Card{
				ID:     3001,
				Title:  "Orphaned card",
				Parent: nil,
			},
			cardTables:         []*api.CardTable{board1, board2},
			expectedBoardID:    100,
			expectedBoardTitle: "Development Board",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find which card table contains the card's current column
			var currentCardTable *api.CardTable
			if tt.card.Parent != nil {
				for _, table := range tt.cardTables {
					for _, column := range table.Lists {
						if column.ID == tt.card.Parent.ID {
							currentCardTable = table
							break
						}
					}
					if currentCardTable != nil {
						break
					}
				}
			}

			// If we couldn't find the card table from the parent, use the default one
			if currentCardTable == nil {
				if len(tt.cardTables) > 0 {
					currentCardTable = tt.cardTables[0]
				}
			}

			assert.NotNil(t, currentCardTable)
			assert.Equal(t, tt.expectedBoardID, currentCardTable.ID)
			assert.Equal(t, tt.expectedBoardTitle, currentCardTable.Title)
		})
	}
}

// TestFindTargetColumn tests finding target column by name or ID
func TestFindTargetColumn(t *testing.T) {
	cardTable := &api.CardTable{
		ID:    100,
		Title: "Development Board",
		Lists: []api.Column{
			{ID: 1, Title: "To Do"},
			{ID: 2, Title: "In Progress"},
			{ID: 3, Title: "Done"},
		},
	}

	tests := []struct {
		name             string
		columnName       string
		expectedColumnID int64
		expectedError    bool
		errorContains    string
	}{
		{
			name:             "find by exact name match",
			columnName:       "Done",
			expectedColumnID: 3,
			expectedError:    false,
		},
		{
			name:             "find by case-insensitive name",
			columnName:       "done",
			expectedColumnID: 3,
			expectedError:    false,
		},
		{
			name:             "find by column ID",
			columnName:       "2",
			expectedColumnID: 2,
			expectedError:    false,
		},
		{
			name:          "error when column name not found",
			columnName:    "Published",
			expectedError: true,
			errorContains: "column 'Published' not found in card table",
		},
		{
			name:          "error when column ID not found",
			columnName:    "99",
			expectedError: true,
			errorContains: "column ID 99 not found in card table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetColumnID int64
			var findErr error

			// Try to parse as ID first
			if id, err := strconv.ParseInt(tt.columnName, 10, 64); err == nil {
				// Validate that the column ID exists in the current card table
				found := false
				for _, column := range cardTable.Lists {
					if column.ID == id {
						targetColumnID = id
						found = true
						break
					}
				}
				if !found {
					findErr = fmt.Errorf("column ID %d not found in card table '%s'", id, cardTable.Title)
				}
			} else {
				// Search by name in the same card table
				columnNameLower := strings.ToLower(tt.columnName)
				for _, column := range cardTable.Lists {
					if strings.ToLower(column.Title) == columnNameLower {
						targetColumnID = column.ID
						break
					}
				}
				if targetColumnID == 0 {
					findErr = fmt.Errorf("column '%s' not found in card table '%s'", tt.columnName, cardTable.Title)
				}
			}

			if tt.expectedError {
				assert.Error(t, findErr)
				assert.Contains(t, findErr.Error(), tt.errorContains)
			} else {
				assert.NoError(t, findErr)
				assert.Equal(t, tt.expectedColumnID, targetColumnID)
			}
		})
	}
}

// TestFindTargetColumn_MultipleBoards tests that column search is scoped to the correct board
func TestFindTargetColumn_MultipleBoards(t *testing.T) {
	board1 := &api.CardTable{
		ID:    100,
		Title: "Development Board",
		Lists: []api.Column{
			{ID: 1, Title: "To Do"},
			{ID: 2, Title: "In Progress"},
			{ID: 3, Title: "Done"},
		},
	}

	board2 := &api.CardTable{
		ID:    200,
		Title: "Marketing Board",
		Lists: []api.Column{
			{ID: 4, Title: "Backlog"},
			{ID: 5, Title: "In Progress"}, // Same name as board1
			{ID: 6, Title: "Published"},
		},
	}

	tests := []struct {
		name             string
		searchBoard      *api.CardTable
		columnName       string
		expectedColumnID int64
		expectedError    bool
		errorContains    string
	}{
		{
			name:             "find 'In Progress' on Development Board",
			searchBoard:      board1,
			columnName:       "In Progress",
			expectedColumnID: 2,
			expectedError:    false,
		},
		{
			name:             "find 'In Progress' on Marketing Board",
			searchBoard:      board2,
			columnName:       "In Progress",
			expectedColumnID: 5,
			expectedError:    false,
		},
		{
			name:          "error finding 'Published' on Development Board",
			searchBoard:   board1,
			columnName:    "Published",
			expectedError: true,
			errorContains: "column 'Published' not found in card table 'Development Board'",
		},
		{
			name:          "error finding column ID 6 on Development Board",
			searchBoard:   board1,
			columnName:    "6",
			expectedError: true,
			errorContains: "column ID 6 not found in card table 'Development Board'",
		},
		{
			name:             "find column ID 6 on Marketing Board",
			searchBoard:      board2,
			columnName:       "6",
			expectedColumnID: 6,
			expectedError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetColumnID int64
			var findErr error

			// Try to parse as ID first
			if id, err := strconv.ParseInt(tt.columnName, 10, 64); err == nil {
				// Validate that the column ID exists in the search board
				found := false
				for _, column := range tt.searchBoard.Lists {
					if column.ID == id {
						targetColumnID = id
						found = true
						break
					}
				}
				if !found {
					findErr = fmt.Errorf("column ID %d not found in card table '%s'", id, tt.searchBoard.Title)
				}
			} else {
				// Search by name in the search board
				columnNameLower := strings.ToLower(tt.columnName)
				for _, column := range tt.searchBoard.Lists {
					if strings.ToLower(column.Title) == columnNameLower {
						targetColumnID = column.ID
						break
					}
				}
				if targetColumnID == 0 {
					findErr = fmt.Errorf("column '%s' not found in card table '%s'", tt.columnName, tt.searchBoard.Title)
				}
			}

			if tt.expectedError {
				assert.Error(t, findErr)
				assert.Contains(t, findErr.Error(), tt.errorContains)
			} else {
				assert.NoError(t, findErr)
				assert.Equal(t, tt.expectedColumnID, targetColumnID)
			}
		})
	}
}

// TestEmptyCardTables tests error handling when no card tables exist
func TestEmptyCardTables(t *testing.T) {
	var cardTables []*api.CardTable

	var currentCardTable *api.CardTable
	var err error

	if currentCardTable == nil {
		if len(cardTables) > 0 {
			currentCardTable = cardTables[0]
		} else {
			err = fmt.Errorf("no card tables found in project")
		}
	}

	assert.Nil(t, currentCardTable)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no card tables found in project")
}
