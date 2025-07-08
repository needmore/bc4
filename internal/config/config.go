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
	ClientID       string                    `json:"client_id,omitempty"`
	ClientSecret   string                    `json:"client_secret,omitempty"`
	DefaultAccount string                    `json:"default_account,omitempty"`
	DefaultProject string                    `json:"default_project,omitempty"`
	Accounts       map[string]AccountConfig  `json:"accounts,omitempty"`
	Preferences    PreferencesConfig         `json:"preferences,omitempty"`
}

// AccountConfig represents per-account configuration
type AccountConfig struct {
	Name           string `json:"name"`
	DefaultProject string `json:"default_project,omitempty"`
}

// PreferencesConfig represents user preferences
type PreferencesConfig struct {
	Editor string `json:"editor,omitempty"`
	Pager  string `json:"pager,omitempty"`
	Color  string `json:"color,omitempty"`
}

var configPath string

func init() {
	configDir, _ := os.UserConfigDir()
	configPath = filepath.Join(configDir, "bc4", "config.json")
}

// Load loads the configuration from file
func Load() (*Config, error) {
	// Set environment variable bindings
	viper.SetEnvPrefix("BC4")
	viper.AutomaticEnv()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return empty config for first run
		return &Config{
			Accounts:    make(map[string]AccountConfig),
			Preferences: PreferencesConfig{
				Editor: os.Getenv("EDITOR"),
				Pager:  "less",
				Color:  "auto",
			},
		}, nil
	}

	// Read config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Override with environment variables
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

	// Write config file
	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
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
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		return false
	}

	// Check for auth file
	authDir, _ := os.UserConfigDir()
	authPath := filepath.Join(authDir, "bc4", "auth.json")
	if _, err := os.Stat(authPath); !os.IsNotExist(err) {
		return false
	}

	return true
}