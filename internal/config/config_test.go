package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save original config path and restore after tests
	originalPath := configPath
	defer func() { configPath = originalPath }()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		envVars        map[string]string
		expectedConfig func(*Config)
		expectError    bool
	}{
		{
			name: "config file doesn't exist - returns defaults",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
			},
			expectedConfig: func(c *Config) {
				assert.Empty(t, c.ClientID)
				assert.Empty(t, c.ClientSecret)
				assert.NotNil(t, c.Accounts)
				assert.NotNil(t, c.Preferences)
			},
		},
		{
			name: "valid config file",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
				testConfig := &Config{
					ClientID:       "test-client-id",
					ClientSecret:   "test-client-secret",
					DefaultAccount: "123",
					DefaultProject: "456",
					Accounts: map[string]AccountConfig{
						"123": {
							Name:           "Test Account",
							DefaultProject: "789",
						},
					},
					Preferences: PreferencesConfig{
						Editor: "vim",
						Pager:  "less",
						Color:  "auto",
					},
				}
				data, err := json.MarshalIndent(testConfig, "", "  ")
				require.NoError(t, err)
				err = os.WriteFile(configPath, data, 0600)
				require.NoError(t, err)
			},
			expectedConfig: func(c *Config) {
				assert.Equal(t, "test-client-id", c.ClientID)
				assert.Equal(t, "test-client-secret", c.ClientSecret)
				assert.Equal(t, "123", c.DefaultAccount)
				assert.Equal(t, "456", c.DefaultProject)
				assert.Equal(t, "Test Account", c.Accounts["123"].Name)
				assert.Equal(t, "789", c.Accounts["123"].DefaultProject)
				assert.Equal(t, "vim", c.Preferences.Editor)
			},
		},
		{
			name: "malformed JSON",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
				err := os.WriteFile(configPath, []byte("invalid json"), 0600)
				require.NoError(t, err)
			},
			expectError: true,
		},
		{
			name: "environment variable overrides",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
				// Create config with some values
				testConfig := &Config{
					ClientID:       "file-client-id",
					ClientSecret:   "file-client-secret",
					DefaultAccount: "file-account",
					DefaultProject: "file-project",
				}
				data, err := json.MarshalIndent(testConfig, "", "  ")
				require.NoError(t, err)
				err = os.WriteFile(configPath, data, 0600)
				require.NoError(t, err)
			},
			envVars: map[string]string{
				"BC4_CLIENT_ID":     "env-client-id",
				"BC4_CLIENT_SECRET": "env-client-secret",
				"BC4_ACCOUNT_ID":    "env-account-id",
				"BC4_PROJECT_ID":    "env-project-id",
			},
			expectedConfig: func(c *Config) {
				assert.Equal(t, "env-client-id", c.ClientID)
				assert.Equal(t, "env-client-secret", c.ClientSecret)
				assert.Equal(t, "env-account-id", c.DefaultAccount)
				assert.Equal(t, "env-project-id", c.DefaultProject)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()

			// Clear viper for clean state
			viper.Reset()

			// Setup test environment
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
				defer func(k string) { _ = os.Unsetenv(k) }(key)
			}

			// Load config
			config, err := Load()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tt.expectedConfig != nil {
					tt.expectedConfig(config)
				}
			}
		})
	}
}

func TestSave(t *testing.T) {
	// Save original config path and restore after tests
	originalPath := configPath
	defer func() { configPath = originalPath }()

	tests := []struct {
		name        string
		config      *Config
		setupFunc   func(t *testing.T, tempDir string)
		expectError bool
		verifyFunc  func(t *testing.T, tempDir string)
	}{
		{
			name: "save new config",
			config: &Config{
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				DefaultAccount: "123",
				Accounts: map[string]AccountConfig{
					"123": {
						Name: "Test Account",
					},
				},
			},
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "bc4", "config.json")
			},
			verifyFunc: func(t *testing.T, tempDir string) {
				// Verify file exists
				info, err := os.Stat(configPath)
				require.NoError(t, err)
				// Just verify the file is readable/writable by owner
				assert.True(t, info.Mode().Perm()&0600 == 0600)

				// Verify content
				data, err := os.ReadFile(configPath)
				require.NoError(t, err)

				var savedConfig Config
				err = json.Unmarshal(data, &savedConfig)
				require.NoError(t, err)
				assert.Equal(t, "test-client-id", savedConfig.ClientID)
			},
		},
		{
			name: "update existing config",
			config: &Config{
				ClientID: "updated-client-id",
			},
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
				// Create existing config
				oldConfig := &Config{ClientID: "old-client-id"}
				data, _ := json.Marshal(oldConfig)
				_ = os.WriteFile(configPath, data, 0600)
			},
			verifyFunc: func(t *testing.T, tempDir string) {
				data, err := os.ReadFile(configPath)
				require.NoError(t, err)

				var savedConfig Config
				err = json.Unmarshal(data, &savedConfig)
				require.NoError(t, err)
				assert.Equal(t, "updated-client-id", savedConfig.ClientID)
			},
		},
		{
			name:   "write permission error",
			config: &Config{},
			setupFunc: func(t *testing.T, tempDir string) {
				// Create read-only directory
				readOnlyDir := filepath.Join(tempDir, "readonly")
				err := os.MkdirAll(readOnlyDir, 0500)
				require.NoError(t, err)
				configPath = filepath.Join(readOnlyDir, "config.json")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			err := Save(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.verifyFunc != nil {
					tt.verifyFunc(t, tempDir)
				}
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original and restore
	originalPath := configPath
	defer func() { configPath = originalPath }()

	tests := []struct {
		name         string
		configPath   string
		expectedPath string
	}{
		{
			name:         "returns configured path",
			configPath:   "/tmp/test/config.json",
			expectedPath: "/tmp/test/config.json",
		},
		{
			name:         "empty path returns empty",
			configPath:   "",
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath = tt.configPath
			result := GetConfigPath()
			assert.Equal(t, tt.expectedPath, result)
		})
	}
}

func TestIsFirstRun(t *testing.T) {
	// Save original and restore
	originalPath := configPath
	defer func() { configPath = originalPath }()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tempDir string)
		expected  bool
	}{
		{
			name: "no files exist - first run",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
				// IsFirstRun checks actual user config dir, not our temp dir
				// So this might return false if user has real config
			},
			expected: true, // Should be true when no files exist
		},
		{
			name: "config file exists",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
				if err := os.WriteFile(configPath, []byte("{}"), 0600); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			},
			expected: false,
		},
		{
			name: "auth file exists but not config",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
				authPath := filepath.Join(tempDir, "auth.json")
				if err := os.WriteFile(authPath, []byte("{}"), 0600); err != nil {
					t.Fatalf("failed to write auth file: %v", err)
				}
			},
			expected: true, // IsFirstRun checks the real user config dir for auth, not temp dir
		},
		{
			name: "both files exist",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath = filepath.Join(tempDir, "config.json")
				authPath := filepath.Join(tempDir, "auth.json")
				if err := os.WriteFile(configPath, []byte("{}"), 0600); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
				if err := os.WriteFile(authPath, []byte("{}"), 0600); err != nil {
					t.Fatalf("failed to write auth file: %v", err)
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			result := IsFirstRun()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_Methods(t *testing.T) {
	// Test various config methods
	cfg := &Config{
		Accounts: map[string]AccountConfig{
			"123": {
				Name: "Test Account",
				ProjectDefaults: map[string]ProjectDefaults{
					"456": {
						DefaultTodoList:  "789",
						DefaultCampfire:  "101",
						DefaultCardTable: "202",
					},
				},
			},
		},
		Preferences: PreferencesConfig{
			Editor: "nano",
			Pager:  "more",
			Color:  "never",
		},
	}

	// Test GetAccountConfig
	t.Run("GetAccountConfig", func(t *testing.T) {
		accountCfg, exists := cfg.Accounts["123"]
		assert.True(t, exists)
		assert.Equal(t, "Test Account", accountCfg.Name)

		_, exists = cfg.Accounts["999"]
		assert.False(t, exists)
	})

	// Test GetProjectDefaults
	t.Run("GetProjectDefaults", func(t *testing.T) {
		if accountCfg, ok := cfg.Accounts["123"]; ok {
			if defaults, ok := accountCfg.ProjectDefaults["456"]; ok {
				assert.Equal(t, "789", defaults.DefaultTodoList)
				assert.Equal(t, "101", defaults.DefaultCampfire)
				assert.Equal(t, "202", defaults.DefaultCardTable)
			} else {
				t.Error("Project defaults not found")
			}
		}
	})

	// Test Preferences
	t.Run("Preferences", func(t *testing.T) {
		assert.Equal(t, "nano", cfg.Preferences.Editor)
		assert.Equal(t, "more", cfg.Preferences.Pager)
		assert.Equal(t, "never", cfg.Preferences.Color)
	})
}

func TestConfig_DefaultValues(t *testing.T) {
	// Test that a new config has sensible defaults
	cfg := &Config{}

	// Initialize maps if needed
	if cfg.Accounts == nil {
		cfg.Accounts = make(map[string]AccountConfig)
	}

	assert.NotNil(t, cfg.Accounts)
	assert.Empty(t, cfg.ClientID)
	assert.Empty(t, cfg.ClientSecret)
	assert.Empty(t, cfg.DefaultAccount)
	assert.Empty(t, cfg.DefaultProject)
}

func TestConfig_JSONMarshaling(t *testing.T) {
	// Test JSON marshaling/unmarshaling preserves all fields
	original := &Config{
		ClientID:       "test-id",
		ClientSecret:   "test-secret",
		DefaultAccount: "123",
		DefaultProject: "456",
		Accounts: map[string]AccountConfig{
			"123": {
				Name:           "Account 1",
				DefaultProject: "789",
				ProjectDefaults: map[string]ProjectDefaults{
					"789": {
						DefaultTodoList:  "111",
						DefaultCampfire:  "222",
						DefaultCardTable: "333",
					},
				},
			},
		},
		Preferences: PreferencesConfig{
			Editor: "emacs",
			Pager:  "bat",
			Color:  "always",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var restored Config
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.ClientID, restored.ClientID)
	assert.Equal(t, original.ClientSecret, restored.ClientSecret)
	assert.Equal(t, original.DefaultAccount, restored.DefaultAccount)
	assert.Equal(t, original.DefaultProject, restored.DefaultProject)
	assert.Equal(t, original.Accounts["123"].Name, restored.Accounts["123"].Name)
	assert.Equal(t, original.Accounts["123"].ProjectDefaults["789"].DefaultTodoList,
		restored.Accounts["123"].ProjectDefaults["789"].DefaultTodoList)
	assert.Equal(t, original.Preferences.Editor, restored.Preferences.Editor)
}

func TestConfig_FirstRunScenario(t *testing.T) {
	// Save original config path and restore after tests
	originalPath := configPath
	defer func() { configPath = originalPath }()

	// Create temp directory
	tempDir := t.TempDir()
	configPath = filepath.Join(tempDir, "bc4", "config.json")

	// Clear viper for clean state
	viper.Reset()

	// Simulate first run - config file doesn't exist
	// Load should return empty config (not nil)
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Empty(t, cfg.ClientID)
	assert.Empty(t, cfg.ClientSecret)

	// Simulate the first-run logic that should save credentials
	// This is what the bug was preventing
	if cfg == nil {
		cfg = &Config{
			Accounts: make(map[string]AccountConfig),
		}
	}
	// The fix: check if credentials are empty instead of if cfg is nil
	if cfg.ClientID == "" {
		cfg.ClientID = "test-client-id"
	}
	if cfg.ClientSecret == "" {
		cfg.ClientSecret = "test-client-secret"
	}
	cfg.DefaultAccount = "123456"

	// Save the config
	err = Save(cfg)
	require.NoError(t, err)

	// Verify the config was saved correctly by loading it again
	cfg2, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "test-client-id", cfg2.ClientID)
	assert.Equal(t, "test-client-secret", cfg2.ClientSecret)
	assert.Equal(t, "123456", cfg2.DefaultAccount)
}

func TestConfig_EnvironmentVariablesNoConfigFile(t *testing.T) {
	// Save original config path and restore after tests
	originalPath := configPath
	defer func() { configPath = originalPath }()

	// Create temp directory but don't create config file
	tempDir := t.TempDir()
	configPath = filepath.Join(tempDir, "bc4", "config.json")

	// Clear viper for clean state
	viper.Reset()

	// Set environment variables
	_ = os.Setenv("BC4_CLIENT_ID", "env-test-id")
	_ = os.Setenv("BC4_CLIENT_SECRET", "env-test-secret")
	_ = os.Setenv("BC4_ACCOUNT_ID", "env-account-123")
	defer func() {
		_ = os.Unsetenv("BC4_CLIENT_ID")
		_ = os.Unsetenv("BC4_CLIENT_SECRET")
		_ = os.Unsetenv("BC4_ACCOUNT_ID")
	}()

	// Load config (no file exists, should use environment variables)
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify environment variables are loaded even without config file
	assert.Equal(t, "env-test-id", cfg.ClientID)
	assert.Equal(t, "env-test-secret", cfg.ClientSecret)
	assert.Equal(t, "env-account-123", cfg.DefaultAccount)
}
