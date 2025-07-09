package auth

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
)

// NewAuthCmd creates the auth command
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  `Authenticate with Basecamp using OAuth2`,
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newRefreshCmd())

	return cmd
}

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in to Basecamp",
		Long:  `Authenticate with Basecamp using OAuth2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Check if credentials are configured
			if cfg.ClientID == "" || cfg.ClientSecret == "" {
				return fmt.Errorf("OAuth credentials not configured. Run 'bc4' to start the setup wizard")
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

			// Perform login
			fmt.Println("Starting authentication flow...")
			token, err := authClient.Login(context.Background())
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Successfully authenticated with %s", token.AccountName)))
			return nil
		},
	}
}

func newLogoutCmd() *cobra.Command {
	var all bool
	
	cmd := &cobra.Command{
		Use:   "logout [account-id]",
		Short: "Log out of Basecamp",
		Long:  `Remove stored authentication tokens`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

			accountID := ""
			if len(args) > 0 {
				accountID = args[0]
			} else if !all {
				// Use default account
				accountID = authClient.GetDefaultAccount()
			}

			// Logout
			if err := authClient.Logout(accountID); err != nil {
				return err
			}

			if all || accountID == "" {
				fmt.Println(successStyle.Render("✓ Logged out of all accounts"))
			} else {
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ Logged out of account %s", accountID)))
			}
			
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Log out of all accounts")
	return cmd
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  `Display current authentication status and account information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Check if credentials are configured
			if cfg.ClientID == "" || cfg.ClientSecret == "" {
				fmt.Println(errorStyle.Render("✗ OAuth credentials not configured"))
				fmt.Println("\nRun 'bc4' to start the setup wizard")
				return nil
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

			// Get accounts
			accounts := authClient.GetAccounts()
			if len(accounts) == 0 {
				fmt.Println(errorStyle.Render("✗ Not authenticated"))
				fmt.Println("\nRun 'bc4 auth login' to authenticate")
				return nil
			}

			// Display status
			fmt.Println(successStyle.Render("✓ Authenticated"))
			fmt.Println()
			
			defaultAccount := authClient.GetDefaultAccount()
			fmt.Println(infoStyle.Render("Accounts:"))
			for _, account := range accounts {
				prefix := "  "
				if account.AccountID == defaultAccount {
					prefix = "• "
				}
				fmt.Printf("%s%s (ID: %s)\n", prefix, account.AccountName, account.AccountID)
			}

			if defaultAccount != "" {
				fmt.Println()
				fmt.Println(infoStyle.Render("Default account: ") + defaultAccount)
			}

			return nil
		},
	}
}

func newRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh [account-id]",
		Short: "Refresh authentication token",
		Long:  `Manually refresh the authentication token for an account`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

			accountID := ""
			if len(args) > 0 {
				accountID = args[0]
			}

			// Get token (will auto-refresh if needed)
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return err
			}

			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Token refreshed for %s", token.AccountName)))
			return nil
		},
	}
}

// GetAuthClient creates an authenticated client from the current configuration
func GetAuthClient() (*auth.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("OAuth credentials not configured")
	}

	return auth.NewClient(cfg.ClientID, cfg.ClientSecret), nil
}