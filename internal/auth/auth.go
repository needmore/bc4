package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"

	"github.com/needmore/bc4/internal/version"
)

const (
	authURL      = "https://launchpad.37signals.com/authorization/new"
	tokenURL     = "https://launchpad.37signals.com/authorization/token"
	callbackPort = "8888"
	redirectURL  = "http://localhost:" + callbackPort + "/callback"
)

// TokenData represents the stored OAuth token information
type TokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ObtainedAt   time.Time `json:"obtained_at"`
}

// AccountToken represents token data for a specific account
type AccountToken struct {
	AccountID    string    `json:"account_id"`
	AccountName  string    `json:"account_name"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ObtainedAt   time.Time `json:"obtained_at"`
}

// AuthStore manages authentication tokens
type AuthStore struct {
	DefaultAccount string                  `json:"default_account"`
	Accounts       map[string]AccountToken `json:"accounts"`
}

// Client handles OAuth2 authentication
type Client struct {
	clientID     string
	clientSecret string
	config       *oauth2.Config
	authStore    *AuthStore
	storePath    string
}

// NewClient creates a new auth client
func NewClient(clientID, clientSecret string) *Client {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirectURL,
		Scopes:      []string{},
	}

	configDir, _ := os.UserConfigDir()
	storePath := filepath.Join(configDir, "bc4", "auth.json")

	client := &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		config:       config,
		storePath:    storePath,
	}

	client.loadAuthStore()
	return client
}

// Login performs the OAuth2 authentication flow
func (c *Client) Login(ctx context.Context) (*AccountToken, error) {
	// Generate state for CSRF protection
	state := c.generateState()

	// Start local HTTP server for callback
	codeChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	server := c.startCallbackServer(state, codeChan, errorChan)
	defer func() {
		if err := server.Shutdown(ctx); err != nil {
			// Log shutdown error but don't fail the operation
			// since we're already in a defer
			_ = err // Explicitly ignore the error
		}
	}()

	// Open browser to authorization URL
	// Basecamp requires a 'type' parameter
	authURL := c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	// Add the required 'type' parameter for Basecamp
	authURL = authURL + "&type=web_server"

	// Try to open browser and provide fallback instructions
	browserErr := browser.OpenURL(authURL)
	if browserErr != nil {
		// Browser couldn't open (e.g., remote SSH session)
		fmt.Println("\nCouldn't open browser automatically.")
		fmt.Println("Please open the following URL in your browser:")
		fmt.Println()
		fmt.Println(authURL)
		fmt.Println("\nWaiting for authentication (Ctrl+C to cancel)...")
	} else {
		fmt.Println("Opening browser for authentication...")
		fmt.Println("If the browser doesn't open, visit this URL:")
		fmt.Println()
		fmt.Println(authURL)
		fmt.Println("\nWaiting for authentication (Ctrl+C to cancel)...")
	}

	// Create a timeout context (5 minutes should be plenty)
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Wait for callback
	select {
	case code := <-codeChan:
		// Exchange code for token
		// Basecamp requires 'type' parameter for token exchange
		token, err := c.config.Exchange(ctx, code,
			oauth2.SetAuthURLParam("type", "web_server"))
		if err != nil {
			return nil, fmt.Errorf("failed to exchange code: %w", err)
		}

		// Create account token
		accountToken := &AccountToken{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			TokenType:    token.TokenType,
			ExpiresIn:    int(time.Until(token.Expiry).Seconds()),
			ObtainedAt:   time.Now(),
		}

		// Get account info and save token
		if err := c.fetchAndSaveAccountInfo(ctx, accountToken); err != nil {
			return nil, err
		}

		return accountToken, nil

	case err := <-errorChan:
		return nil, fmt.Errorf("callback error: %w", err)

	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("authentication timed out after 5 minutes")
		}
		return nil, fmt.Errorf("authentication cancelled")

	case <-ctx.Done():
		return nil, fmt.Errorf("authentication cancelled")
	}
}

// Logout removes stored credentials
func (c *Client) Logout(accountID string) error {
	if c.authStore == nil {
		return nil
	}

	if accountID == "" {
		// Clear all accounts
		c.authStore = &AuthStore{
			Accounts: make(map[string]AccountToken),
		}
	} else {
		// Clear specific account
		delete(c.authStore.Accounts, accountID)
		if c.authStore.DefaultAccount == accountID {
			c.authStore.DefaultAccount = ""
		}
	}

	return c.saveAuthStore()
}

// GetToken returns a valid token for the specified account
func (c *Client) GetToken(accountID string) (*AccountToken, error) {
	if c.authStore == nil || c.authStore.Accounts == nil {
		return nil, fmt.Errorf("no authenticated accounts")
	}

	// Use default account if none specified
	if accountID == "" {
		accountID = c.authStore.DefaultAccount
	}

	token, exists := c.authStore.Accounts[accountID]
	if !exists {
		return nil, fmt.Errorf("account %s not found", accountID)
	}

	// Check if token is expired
	if c.isTokenExpired(&token) {
		// Refresh token
		refreshed, err := c.refreshToken(&token)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		token = *refreshed
		c.authStore.Accounts[accountID] = token
		_ = c.saveAuthStore()
	}

	return &token, nil
}

// GetAccounts returns all authenticated accounts
func (c *Client) GetAccounts() map[string]AccountToken {
	if c.authStore == nil || c.authStore.Accounts == nil {
		return make(map[string]AccountToken)
	}
	return c.authStore.Accounts
}

// GetDefaultAccount returns the default account ID
func (c *Client) GetDefaultAccount() string {
	if c.authStore == nil {
		return ""
	}
	return c.authStore.DefaultAccount
}

// SetDefaultAccount sets the default account
func (c *Client) SetDefaultAccount(accountID string) error {
	if c.authStore == nil {
		c.authStore = &AuthStore{
			Accounts: make(map[string]AccountToken),
		}
	}

	if _, exists := c.authStore.Accounts[accountID]; !exists {
		return fmt.Errorf("account %s not found", accountID)
	}

	c.authStore.DefaultAccount = accountID
	return c.saveAuthStore()
}

// Private methods

func (c *Client) generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (c *Client) startCallbackServer(state string, codeChan chan<- string, errorChan chan<- error) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Verify state
		if r.URL.Query().Get("state") != state {
			errorChan <- fmt.Errorf("invalid state parameter")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			errorChan <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		// Send success response
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `
			<html>
			<head><title>Authentication Successful</title></head>
			<body>
				<h1>Authentication Successful!</h1>
				<p>You can now close this window and return to the terminal.</p>
				<script>window.close();</script>
			</body>
			</html>
		`)

		codeChan <- code
	})

	server := &http.Server{
		Addr:    ":" + callbackPort,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errorChan <- err
		}
	}()

	return server
}

func (c *Client) isTokenExpired(token *AccountToken) bool {
	expiresAt := token.ObtainedAt.Add(time.Duration(token.ExpiresIn) * time.Second)
	return time.Now().After(expiresAt.Add(-5 * time.Minute)) // 5 minute buffer
}

func (c *Client) refreshToken(token *AccountToken) (*AccountToken, error) {
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Create form data
	data := url.Values{}
	data.Set("type", "refresh")
	data.Set("refresh_token", token.RefreshToken)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: %s", resp.Status)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Update token
	token.AccessToken = result.AccessToken
	if result.RefreshToken != "" {
		token.RefreshToken = result.RefreshToken
	}
	token.ExpiresIn = result.ExpiresIn
	token.ObtainedAt = time.Now()

	return token, nil
}

func (c *Client) fetchAndSaveAccountInfo(ctx context.Context, token *AccountToken) error {
	// Get authorization info to find account ID
	req, err := http.NewRequestWithContext(ctx, "GET", "https://launchpad.37signals.com/authorization.json", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("User-Agent", version.UserAgent())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var authInfo struct {
		Accounts []struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			Product string `json:"product"`
		} `json:"accounts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authInfo); err != nil {
		return err
	}

	// Silent - no debug output during auth

	// Save token(s) for all Basecamp accounts found
	if c.authStore == nil {
		c.authStore = &AuthStore{
			Accounts: make(map[string]AccountToken),
		}
	}

	basecampAccounts := 0
	firstAccountID := ""

	// Save all BC3/BC4 accounts
	for _, account := range authInfo.Accounts {
		if account.Product == "bc3" || account.Product == "bc4" || account.Product == "basecamp3" || account.Product == "basecamp4" || account.Product == "basecamp" {
			accountToken := AccountToken{
				AccountID:    fmt.Sprintf("%d", account.ID),
				AccountName:  account.Name,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				TokenType:    token.TokenType,
				ExpiresIn:    token.ExpiresIn,
				ObtainedAt:   token.ObtainedAt,
			}

			c.authStore.Accounts[accountToken.AccountID] = accountToken
			basecampAccounts++

			if firstAccountID == "" {
				firstAccountID = accountToken.AccountID
			}

			// Silent - added account
		}
	}

	if basecampAccounts == 0 {
		return fmt.Errorf("no Basecamp accounts found")
	}

	// Set default account if not set
	if c.authStore.DefaultAccount == "" {
		c.authStore.DefaultAccount = firstAccountID
	}

	// Update the token to return with the first account info
	if basecampAccounts > 0 {
		firstToken := c.authStore.Accounts[firstAccountID]
		token.AccountID = firstToken.AccountID
		token.AccountName = firstToken.AccountName
	}

	return c.saveAuthStore()
}

func (c *Client) loadAuthStore() {
	file, err := os.Open(c.storePath)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	c.authStore = &AuthStore{}
	_ = json.NewDecoder(file).Decode(c.authStore)
}

func (c *Client) saveAuthStore() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(c.storePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Write file with restricted permissions
	file, err := os.OpenFile(c.storePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c.authStore)
}
