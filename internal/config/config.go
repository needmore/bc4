package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	AccountID    string `mapstructure:"account_id"`
	AccessToken  string `mapstructure:"access_token"`
	BaseURL      string `mapstructure:"base_url"`
}

func Load() (*Config, error) {
	cfg := &Config{
		BaseURL: "https://3.basecamp.com",
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.AccountID == "" {
		return nil, fmt.Errorf("account_id is required - set via config file, BC4_ACCOUNT_ID env var, or --account-id flag")
	}

	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("access_token is required - set via config file, BC4_ACCESS_TOKEN env var, or --access-token flag")
	}

	return cfg, nil
}

func SaveExample() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	example := `# Basecamp 4 CLI Configuration
# 
# Get your account ID from your Basecamp URL: https://3.basecamp.com/[ACCOUNT_ID]
# Generate a personal access token at: https://3.basecamp.com/[ACCOUNT_ID]/my/profile

account_id: "YOUR_ACCOUNT_ID"
access_token: "YOUR_ACCESS_TOKEN"

# Optional: Override the base URL (default: https://3.basecamp.com)
# base_url: "https://3.basecamp.com"
`

	configPath := fmt.Sprintf("%s/.bc4.yaml.example", home)
	return os.WriteFile(configPath, []byte(example), 0644)
}