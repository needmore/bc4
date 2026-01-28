package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	ClientID       string                   `json:"client_id,omitempty"`
	ClientSecret   string                   `json:"client_secret,omitempty"`
	DefaultAccount string                   `json:"default_account,omitempty"`
	DefaultProject string                   `json:"default_project,omitempty"`
	Accounts       map[string]AccountConfig `json:"accounts,omitempty"`
	Preferences    PreferencesConfig        `json:"preferences,omitempty"`
}

// AccountConfig represents per-account configuration
type AccountConfig struct {
	Name            string                     `json:"name"`
	DefaultProject  string                     `json:"default_project,omitempty"`
	ProjectDefaults map[string]ProjectDefaults `json:"project_defaults,omitempty"`
}

// ProjectDefaults represents per-project default settings
type ProjectDefaults struct {
	DefaultTodoList  string `json:"default_todo_list,omitempty"`
	DefaultCampfire  string `json:"default_campfire,omitempty"`
	DefaultCardTable string `json:"default_card_table,omitempty"`
}

// PreferencesConfig represents user preferences
type PreferencesConfig struct {
	Editor string `json:"editor,omitempty"`
	Pager  string `json:"pager,omitempty"`
	Color  string `json:"color,omitempty"`
}

var configPath string
var authPath string

func init() {
	configDir, _ := os.UserConfigDir()
	configPath = filepath.Join(configDir, "bc4", "config.json")
	authPath = filepath.Join(configDir, "bc4", "auth.json")
}

// Load loads the configuration from file
func Load() (*Config, error) {
	// Set environment variable bindings
	viper.SetEnvPrefix("BC4")
	viper.AutomaticEnv()

	var config Config

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return empty config for first run
		config = Config{
			Accounts: make(map[string]AccountConfig),
			Preferences: PreferencesConfig{
				Editor: os.Getenv("EDITOR"),
				Pager:  "less",
				Color:  "auto",
			},
		}
	} else {
		// Read config file
		file, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
		defer func() { _ = file.Close() }()

		if err := json.NewDecoder(file).Decode(&config); err != nil {
			return nil, fmt.Errorf("failed to decode config: %w", err)
		}
	}

	// Override with environment variables (applies to both file and no-file cases)
	if clientID := viper.GetString("CLIENT_ID"); clientID != "" {
		config.ClientID = clientID
	}
	if clientSecret := viper.GetString("CLIENT_SECRET"); clientSecret != "" {
		config.ClientSecret = clientSecret
	}
	if accountID := viper.GetString("ACCOUNT_ID"); accountID != "" {
		config.DefaultAccount = accountID
	}
	if projectID := viper.GetString("PROJECT_ID"); projectID != "" {
		config.DefaultProject = projectID
	}

	return &config, nil
}

// Save saves the configuration to file
func Save(config *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpFile, err := os.CreateTemp(dir, ".config-*.json.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	tmpPath := tmpFile.Name()

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to encode config: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if err := os.Rename(tmpPath, configPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configPath
}

// IsFirstRun checks if this is the first run (no auth or config)
func IsFirstRun() bool {
	// Check for config file
	if _, err := os.Stat(configPath); err == nil {
		return false
	}

	// Check for auth file
	if _, err := os.Stat(authPath); err == nil {
		return false
	}

	return true
}
