package factory

import (
	"context"
	"fmt"
	"sync"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/errors"
)

// Factory provides centralized dependency management for commands
type Factory struct {
	// Configuration
	config     *config.Config
	configOnce sync.Once
	configErr  error

	// Auth client
	authClient     *auth.Client
	authClientOnce sync.Once

	// API client
	apiClient     *api.ModularClient
	apiClientOnce sync.Once
	apiClientErr  error

	// Override fields for specific scenarios
	accountID string
	projectID string
}

// New creates a new factory instance
func New() *Factory {
	return &Factory{}
}

// WithAccount sets a specific account ID to use
func (f *Factory) WithAccount(accountID string) *Factory {
	f.accountID = accountID
	// Reset API client since it depends on account
	f.apiClient = nil
	f.apiClientOnce = sync.Once{}
	return f
}

// WithProject sets a specific project ID to use
func (f *Factory) WithProject(projectID string) *Factory {
	f.projectID = projectID
	return f
}

// Config returns the loaded configuration, loading it once if needed
func (f *Factory) Config() (*config.Config, error) {
	f.configOnce.Do(func() {
		f.config, f.configErr = config.Load()
	})
	return f.config, f.configErr
}

// AuthClient returns the auth client, creating it once if needed
func (f *Factory) AuthClient() (*auth.Client, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.NewAuthenticationError(fmt.Errorf("not authenticated"))
	}

	f.authClientOnce.Do(func() {
		f.authClient = auth.NewClient(cfg.ClientID, cfg.ClientSecret)
	})

	return f.authClient, nil
}

// AccountID returns the account ID to use, either from override or default
func (f *Factory) AccountID() (string, error) {
	if f.accountID != "" {
		return f.accountID, nil
	}

	authClient, err := f.AuthClient()
	if err != nil {
		return "", err
	}

	accountID := authClient.GetDefaultAccount()
	if accountID == "" {
		return "", errors.NewConfigurationError("no account specified and no default account set", nil)
	}

	return accountID, nil
}

// ProjectID returns the project ID to use, either from override or default
func (f *Factory) ProjectID() (string, error) {
	if f.projectID != "" {
		return f.projectID, nil
	}

	cfg, err := f.Config()
	if err != nil {
		return "", err
	}

	accountID, err := f.AccountID()
	if err != nil {
		return "", err
	}

	projectID := cfg.DefaultProject
	if projectID == "" && cfg.Accounts != nil {
		if acc, ok := cfg.Accounts[accountID]; ok {
			projectID = acc.DefaultProject
		}
	}

	if projectID == "" {
		return "", errors.NewConfigurationError("no project specified and no default project set", nil)
	}

	return projectID, nil
}

// ApiClient returns the API client, creating it once if needed
func (f *Factory) ApiClient() (*api.ModularClient, error) {
	f.apiClientOnce.Do(func() {
		accountID, err := f.AccountID()
		if err != nil {
			f.apiClientErr = err
			return
		}

		authClient, err := f.AuthClient()
		if err != nil {
			f.apiClientErr = err
			return
		}

		token, err := authClient.GetToken(accountID)
		if err != nil {
			f.apiClientErr = fmt.Errorf("failed to get auth token: %w", err)
			return
		}

		f.apiClient = api.NewModularClient(accountID, token.AccessToken)
	})

	return f.apiClient, f.apiClientErr
}

// Context returns a context for API operations
// This can be extended in the future to include timeouts, tracing, etc.
func (f *Factory) Context() context.Context {
	return context.Background()
}
