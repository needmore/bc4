package tableprinter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// snapshotPath returns the path to the snapshot file for a test
func snapshotPath(t *testing.T) string {
	return filepath.Join("testdata", "snapshots", strings.ReplaceAll(t.Name(), "/", "_")+".txt")
}

// updateSnapshots checks if snapshot updates are enabled
func updateSnapshots() bool {
	return os.Getenv("UPDATE_SNAPSHOTS") == "true"
}

// assertSnapshot compares output with a snapshot file
func assertSnapshot(t *testing.T, actual string) {
	t.Helper()

	snapPath := snapshotPath(t)
	snapDir := filepath.Dir(snapPath)

	if updateSnapshots() {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(snapDir, 0755); err != nil {
			t.Fatalf("Failed to create snapshot directory: %v", err)
		}

		// Write the snapshot
		if err := os.WriteFile(snapPath, []byte(actual), 0644); err != nil {
			t.Fatalf("Failed to write snapshot: %v", err)
		}

		t.Logf("Updated snapshot: %s", snapPath)
		return
	}

	// Read the expected snapshot
	expected, err := os.ReadFile(snapPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Snapshot does not exist: %s\nRun with UPDATE_SNAPSHOTS=true to create it", snapPath)
		}
		t.Fatalf("Failed to read snapshot: %v", err)
	}

	// Compare actual with expected
	if string(expected) != actual {
		t.Errorf("Output does not match snapshot.\nExpected:\n%s\nActual:\n%s\n\nRun with UPDATE_SNAPSHOTS=true to update the snapshot",
			string(expected), actual)
	}
}

// TestTableSnapshots tests table output with snapshots
func TestTableSnapshots(t *testing.T) {
	// Color function for testing
	greenColor := func(s string) string {
		return "\033[32m" + s + "\033[0m"
	}
	redColor := func(s string) string {
		return "\033[31m" + s + "\033[0m"
	}

	tests := []struct {
		name       string
		isTTY      bool
		termWidth  int
		setupTable func(TablePrinter)
	}{
		{
			name:      "simple_table",
			isTTY:     true,
			termWidth: 80,
			setupTable: func(table TablePrinter) {
				table.AddHeader([]string{"Name", "Age", "City"})
				table.AddField("Alice")
				table.AddField("30")
				table.AddField("New York")
				table.EndRow()
				table.AddField("Bob")
				table.AddField("25")
				table.AddField("Los Angeles")
				table.EndRow()
				table.AddField("Charlie")
				table.AddField("35")
				table.AddField("Chicago")
				table.EndRow()
			},
		},
		{
			name:      "table_with_colors",
			isTTY:     true,
			termWidth: 80,
			setupTable: func(table TablePrinter) {
				table.AddHeader([]string{"Status", "Count", "Percentage"})
				table.AddField("Success", WithColor(greenColor))
				table.AddField("150")
				table.AddField("75%")
				table.EndRow()
				table.AddField("Failed", WithColor(redColor))
				table.AddField("30")
				table.AddField("15%")
				table.EndRow()
				table.AddField("Pending")
				table.AddField("20")
				table.AddField("10%")
				table.EndRow()
			},
		},
		{
			name:      "non_tty_output",
			isTTY:     false,
			termWidth: 0,
			setupTable: func(table TablePrinter) {
				table.AddHeader([]string{"ID", "Name", "Value"})
				table.AddField("1")
				table.AddField("Item One")
				table.AddField("100")
				table.EndRow()
				table.AddField("2")
				table.AddField("Item Two")
				table.AddField("200")
				table.EndRow()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Create the table
			table := New(&buf, tt.isTTY, tt.termWidth)

			// Setup the table
			tt.setupTable(table)

			// Render the table
			if err := table.Render(); err != nil {
				t.Fatalf("Failed to render table: %v", err)
			}

			// Compare with snapshot
			assertSnapshot(t, buf.String())
		})
	}
}

// TestProjectTableSnapshot tests a real-world project table output
func TestProjectTableSnapshot(t *testing.T) {
	var buf bytes.Buffer

	table := New(&buf, true, 120)
	table.AddHeader([]string{"ID", "Name", "Description", "Created", "Updated"})

	// Add sample project data
	projects := [][]string{
		{"12345", "Web Redesign", "Complete redesign of company website", "2024-01-15", "2024-11-20"},
		{"12346", "Mobile App", "Native iOS and Android apps", "2024-02-01", "2024-11-19"},
		{"12347", "API v2", "REST API version 2.0", "2024-03-10", "2024-11-18"},
		{"12348", "Data Migration", "Migrate from MySQL to PostgreSQL", "2024-04-05", "2024-11-17"},
	}

	for _, project := range projects {
		for _, field := range project {
			table.AddField(field)
		}
		table.EndRow()
	}

	if err := table.Render(); err != nil {
		t.Fatalf("Failed to render table: %v", err)
	}

	assertSnapshot(t, buf.String())
}

// TestTodoTableSnapshot tests a todo list table output
func TestTodoTableSnapshot(t *testing.T) {
	var buf bytes.Buffer

	// Color functions
	grayColor := func(s string) string {
		return "\033[90m" + s + "\033[0m"
	}
	greenColor := func(s string) string {
		return "\033[32m" + s + "\033[0m"
	}
	yellowColor := func(s string) string {
		return "\033[33m" + s + "\033[0m"
	}
	redColor := func(s string) string {
		return "\033[31m" + s + "\033[0m"
	}

	table := New(&buf, true, 100)
	table.AddHeader([]string{"", "Title", "Assignee", "Due Date", "Status"})

	// Add sample todo data with colors
	todos := []struct {
		checkbox    string
		title       string
		assignee    string
		dueDate     string
		status      string
		statusColor func(string) string
	}{
		{"□", "Write unit tests", "Alice", "2024-11-25", "In Progress", yellowColor},
		{"□", "Review PR #123", "Bob", "2024-11-24", "Pending", grayColor},
		{"✓", "Fix bug #456", "Charlie", "2024-11-23", "Completed", greenColor},
		{"□", "Update documentation", "Alice", "2024-11-26", "Not Started", grayColor},
		{"□", "Deploy to staging", "", "2024-11-27", "Blocked", redColor},
	}

	for _, todo := range todos {
		table.AddField(todo.checkbox)
		table.AddField(todo.title)
		table.AddField(todo.assignee)
		table.AddField(todo.dueDate)
		table.AddField(todo.status, WithColor(todo.statusColor))
		table.EndRow()
	}

	if err := table.Render(); err != nil {
		t.Fatalf("Failed to render table: %v", err)
	}

	assertSnapshot(t, buf.String())
}

