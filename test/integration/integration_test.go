//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/config"
)

// testConfig holds test configuration
type testConfig struct {
	AccountID    string
	AccessToken  string
	TestProject  string
	SkipCleanup  bool
}

// getTestConfig loads test configuration from environment
func getTestConfig() (*testConfig, error) {
	cfg := &testConfig{
		AccountID:   os.Getenv("BC4_TEST_ACCOUNT_ID"),
		AccessToken: os.Getenv("BC4_TEST_ACCESS_TOKEN"),
		TestProject: os.Getenv("BC4_TEST_PROJECT_ID"),
		SkipCleanup: os.Getenv("BC4_TEST_SKIP_CLEANUP") == "true",
	}

	if cfg.AccountID == "" || cfg.AccessToken == "" {
		return nil, skipTestError("BC4_TEST_ACCOUNT_ID and BC4_TEST_ACCESS_TOKEN must be set")
	}

	return cfg, nil
}

// skipTestError returns an error that causes the test to be skipped
func skipTestError(msg string) error {
	return &testSkipError{msg: msg}
}

type testSkipError struct {
	msg string
}

func (e *testSkipError) Error() string {
	return e.msg
}

// skipIfNeeded checks if tests should be skipped
func skipIfNeeded(t *testing.T, err error) {
	if _, ok := err.(*testSkipError); ok {
		t.Skip(err.Error())
	}
}

// TestAPIConnection verifies basic API connectivity
func TestAPIConnection(t *testing.T) {
	cfg, err := getTestConfig()
	if err != nil {
		skipIfNeeded(t, err)
		t.Fatal(err)
	}

	client := api.NewModularClient(cfg.AccountID, cfg.AccessToken)
	ctx := context.Background()

	// Try to get projects
	projects, err := client.Projects().GetProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}

	if len(projects) == 0 {
		t.Fatal("Expected at least one project")
	}

	t.Logf("Successfully connected to account %s with %d projects", cfg.AccountID, len(projects))
}

// TestProjectOperations tests project-related operations
func TestProjectOperations(t *testing.T) {
	cfg, err := getTestConfig()
	if err != nil {
		skipIfNeeded(t, err)
		t.Fatal(err)
	}

	client := api.NewModularClient(cfg.AccountID, cfg.AccessToken)
	ctx := context.Background()

	// Get all projects
	projects, err := client.Projects().GetProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}

	if len(projects) == 0 {
		t.Skip("No projects available for testing")
	}

	// Test getting a specific project
	project := projects[0]
	retrieved, err := client.Projects().GetProject(ctx, fmt.Sprintf("%d", project.ID))
	if err != nil {
		t.Fatalf("Failed to get project %d: %v", project.ID, err)
	}

	if retrieved.ID != project.ID {
		t.Errorf("Expected project ID %d, got %d", project.ID, retrieved.ID)
	}

	if retrieved.Name != project.Name {
		t.Errorf("Expected project name %s, got %s", project.Name, retrieved.Name)
	}
}

// TestTodoOperations tests todo-related operations
func TestTodoOperations(t *testing.T) {
	cfg, err := getTestConfig()
	if err != nil {
		skipIfNeeded(t, err)
		t.Fatal(err)
	}

	if cfg.TestProject == "" {
		t.Skip("BC4_TEST_PROJECT_ID not set, skipping todo tests")
	}

	client := api.NewModularClient(cfg.AccountID, cfg.AccessToken)
	ctx := context.Background()

	// Get the todo set for the project
	todoSet, err := client.Todos().GetProjectTodoSet(ctx, cfg.TestProject)
	if err != nil {
		t.Fatalf("Failed to get todo set: %v", err)
	}

	if todoSet == nil {
		t.Skip("Project has no todo set")
	}

	// Get todo lists
	lists, err := client.Todos().GetTodoLists(ctx, cfg.TestProject, todoSet.ID)
	if err != nil {
		t.Fatalf("Failed to get todo lists: %v", err)
	}

	if len(lists) == 0 {
		t.Skip("No todo lists available")
	}

	// Create a test todo
	testList := lists[0]
	req := api.TodoCreateRequest{
		Content:     "Integration test todo",
		Description: "This todo was created by integration tests",
	}

	created, err := client.Todos().CreateTodo(ctx, cfg.TestProject, testList.ID, req)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	t.Logf("Created todo %d: %s", created.ID, created.Content)

	// Clean up if not skipping
	if !cfg.SkipCleanup {
		err = client.Todos().CompleteTodo(ctx, cfg.TestProject, created.ID)
		if err != nil {
			t.Errorf("Failed to complete todo %d: %v", created.ID, err)
		} else {
			t.Logf("Completed todo %d", created.ID)
		}
	}
}

// TestCampfireOperations tests campfire-related operations
func TestCampfireOperations(t *testing.T) {
	cfg, err := getTestConfig()
	if err != nil {
		skipIfNeeded(t, err)
		t.Fatal(err)
	}

	if cfg.TestProject == "" {
		t.Skip("BC4_TEST_PROJECT_ID not set, skipping campfire tests")
	}

	client := api.NewModularClient(cfg.AccountID, cfg.AccessToken)
	ctx := context.Background()

	// List campfires
	campfires, err := client.Campfires().ListCampfires(ctx, cfg.TestProject)
	if err != nil {
		t.Fatalf("Failed to list campfires: %v", err)
	}

	if len(campfires) == 0 {
		t.Skip("No campfires available")
	}

	// Get campfire lines
	campfire := campfires[0]
	lines, err := client.Campfires().GetCampfireLines(ctx, cfg.TestProject, campfire.ID, 10)
	if err != nil {
		t.Fatalf("Failed to get campfire lines: %v", err)
	}

	t.Logf("Campfire %s has %d recent messages", campfire.Name, len(lines))
}

// TestCardTableOperations tests card table operations
func TestCardTableOperations(t *testing.T) {
	cfg, err := getTestConfig()
	if err != nil {
		skipIfNeeded(t, err)
		t.Fatal(err)
	}

	if cfg.TestProject == "" {
		t.Skip("BC4_TEST_PROJECT_ID not set, skipping card table tests")
	}

	client := api.NewModularClient(cfg.AccountID, cfg.AccessToken)
	ctx := context.Background()

	// Get the card table for the project
	cardTable, err := client.Cards().GetProjectCardTable(ctx, cfg.TestProject)
	if err != nil {
		t.Fatalf("Failed to get card table: %v", err)
	}

	if cardTable == nil {
		t.Skip("Project has no card table")
	}

	if len(cardTable.Lists) == 0 {
		t.Skip("Card table has no columns")
	}

	// Get cards in the first column
	column := cardTable.Lists[0]
	cards, err := client.Cards().GetCardsInColumn(ctx, cfg.TestProject, column.ID)
	if err != nil {
		t.Fatalf("Failed to get cards in column %s: %v", column.Title, err)
	}

	t.Logf("Column %s has %d cards", column.Title, len(cards))
}

// TestConfigOperations tests configuration management
func TestConfigOperations(t *testing.T) {
	// Create a temporary config directory
	tmpDir, err := os.MkdirTemp("", "bc4-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set config path to temp directory
	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)

	// Test loading empty config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load empty config: %v", err)
	}

	// Test saving config
	cfg.DefaultAccount = "test-account"
	cfg.Preferences.Editor = "vim"
	
	err = config.Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test loading saved config
	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.DefaultAccount != "test-account" {
		t.Errorf("Expected default account 'test-account', got '%s'", loaded.DefaultAccount)
	}

	if loaded.Preferences.Editor != "vim" {
		t.Errorf("Expected editor 'vim', got '%s'", loaded.Preferences.Editor)
	}
}