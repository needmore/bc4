package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewClient(t *testing.T) {
	// Create a temporary directory for test
	tempDir := t.TempDir()
	
	// Override the config directory for testing
	oldConfigDir := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldConfigDir)
	
	client := NewClient("test-client-id", "test-client-secret")
	assert.NotNil(t, client)
	assert.Equal(t, "test-client-id", client.clientID)
	assert.Equal(t, "test-client-secret", client.clientSecret)
	assert.NotNil(t, client.config)
	assert.Equal(t, "test-client-id", client.config.ClientID)
	assert.Equal(t, "test-client-secret", client.config.ClientSecret)
	assert.Equal(t, redirectURL, client.config.RedirectURL)
}

func TestClient_GetAccounts(t *testing.T) {
	client := createTestClient(t)
	
	// Test with empty store
	accounts := client.GetAccounts()
	assert.Empty(t, accounts)
	
	// Add test accounts
	client.authStore = &AuthStore{
		Accounts: map[string]AccountToken{
			"123": {AccountID: "123", AccountName: "Test Account 1"},
			"456": {AccountID: "456", AccountName: "Test Account 2"},
		},
	}
	
	accounts = client.GetAccounts()
	assert.Len(t, accounts, 2)
	
	// Check that we have the accounts
	_, exists := accounts["123"]
	assert.True(t, exists)
	_, exists = accounts["456"]
	assert.True(t, exists)
}

func TestClient_GetDefaultAccount(t *testing.T) {
	client := createTestClient(t)
	
	// Test with no default
	defaultAccount := client.GetDefaultAccount()
	assert.Empty(t, defaultAccount)
	
	// Set default account
	client.authStore = &AuthStore{
		DefaultAccount: "123",
		Accounts: map[string]AccountToken{
			"123": {AccountID: "123", AccountName: "Test Account"},
		},
	}
	
	defaultAccount = client.GetDefaultAccount()
	assert.Equal(t, "123", defaultAccount)
}

func TestClient_SetDefaultAccount(t *testing.T) {
	client := createTestClient(t)
	
	// Initialize auth store
	client.authStore = &AuthStore{
		Accounts: make(map[string]AccountToken),
	}
	
	// Test with non-existent account
	err := client.SetDefaultAccount("999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account 999 not found")
	
	// Add account and set as default
	client.authStore.Accounts["123"] = AccountToken{
		AccountID:   "123",
		AccountName: "Test Account",
	}
	
	err = client.SetDefaultAccount("123")
	require.NoError(t, err)
	assert.Equal(t, "123", client.authStore.DefaultAccount)
}

func TestClient_Logout(t *testing.T) {
	client := createTestClient(t)
	
	// Setup test data
	client.authStore = &AuthStore{
		DefaultAccount: "123",
		Accounts: map[string]AccountToken{
			"123": {AccountID: "123", AccountName: "Test Account 1"},
			"456": {AccountID: "456", AccountName: "Test Account 2"},
		},
	}
	
	// Save to disk
	err := client.saveAuthStore()
	require.NoError(t, err)
	
	t.Run("logout specific account", func(t *testing.T) {
		err := client.Logout("123")
		require.NoError(t, err)
		
		_, exists := client.authStore.Accounts["123"]
		assert.False(t, exists)
		_, exists = client.authStore.Accounts["456"]
		assert.True(t, exists)
		assert.Empty(t, client.authStore.DefaultAccount) // Default should be cleared
	})
	
	t.Run("logout all accounts", func(t *testing.T) {
		err := client.Logout("")
		require.NoError(t, err)
		
		assert.Empty(t, client.authStore.Accounts)
		assert.Empty(t, client.authStore.DefaultAccount)
	})
}

func TestClient_GetToken(t *testing.T) {
	client := createTestClient(t)
	
	t.Run("account not found", func(t *testing.T) {
		_, err := client.GetToken("999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account 999 not found")
	})
	
	t.Run("valid token", func(t *testing.T) {
		// Add account with valid token
		validToken := AccountToken{
			AccountID:    "123",
			AccountName:  "Test Account",
			AccessToken:  "valid-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    3600,
			ObtainedAt:   time.Now(),
		}
		
		client.authStore = &AuthStore{
			Accounts: map[string]AccountToken{
				"123": validToken,
			},
		}
		
		tokenObj, err := client.GetToken("123")
		require.NoError(t, err)
		assert.Equal(t, "valid-token", tokenObj.AccessToken)
	})
	
	t.Run("expired token needs refresh", func(t *testing.T) {
		// Mock OAuth server for token refresh
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/authorization/token" {
				// Return new token
				response := map[string]interface{}{
					"access_token":  "new-access-token",
					"refresh_token": "new-refresh-token",
					"token_type":    "Bearer",
					"expires_in":    3600,
				}
				json.NewEncoder(w).Encode(response)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()
		
		// Override token endpoint
		client.config.Endpoint.TokenURL = server.URL + "/authorization/token"
		
		// Add account with expired token
		expiredToken := AccountToken{
			AccountID:    "123",
			AccountName:  "Test Account",
			AccessToken:  "expired-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    3600,
			ObtainedAt:   time.Now().Add(-2 * time.Hour), // Expired 2 hours ago
		}
		
		client.authStore = &AuthStore{
			Accounts: map[string]AccountToken{
				"123": expiredToken,
			},
		}
		
		tokenObj, err := client.GetToken("123")
		require.NoError(t, err)
		assert.Equal(t, "new-access-token", tokenObj.AccessToken)
		assert.Equal(t, "new-access-token", client.authStore.Accounts["123"].AccessToken)
	})
}

func TestClient_isTokenExpired(t *testing.T) {
	client := &Client{}
	
	tests := []struct {
		name     string
		token    *AccountToken
		expected bool
	}{
		{
			name: "valid token",
			token: &AccountToken{
				ExpiresIn:  3600,
				ObtainedAt: time.Now(),
			},
			expected: false,
		},
		{
			name: "expired token",
			token: &AccountToken{
				ExpiresIn:  3600,
				ObtainedAt: time.Now().Add(-2 * time.Hour),
			},
			expected: true,
		},
		{
			name: "token expiring soon",
			token: &AccountToken{
				ExpiresIn:  3600,
				ObtainedAt: time.Now().Add(-55 * time.Minute), // 5 minutes before expiry
			},
			expected: true, // Should refresh if less than 10 minutes remaining
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isTokenExpired(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_loadAuthStore(t *testing.T) {
	client := createTestClient(t)
	
	t.Run("no auth file exists", func(t *testing.T) {
		client.loadAuthStore()
		assert.NotNil(t, client.authStore)
		assert.Empty(t, client.authStore.Accounts)
	})
	
	t.Run("valid auth file", func(t *testing.T) {
		// Create test auth data
		testStore := &AuthStore{
			DefaultAccount: "123",
			Accounts: map[string]AccountToken{
				"123": {
					AccountID:   "123",
					AccountName: "Test Account",
					AccessToken: "test-token",
				},
			},
		}
		
		// Ensure directory exists
		dir := filepath.Dir(client.storePath)
		err := os.MkdirAll(dir, 0700)
		require.NoError(t, err)
		
		// Save to file
		data, err := json.MarshalIndent(testStore, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(client.storePath, data, 0600)
		require.NoError(t, err)
		
		// Load and verify
		client.loadAuthStore()
		assert.Equal(t, "123", client.authStore.DefaultAccount)
		assert.Len(t, client.authStore.Accounts, 1)
		assert.Equal(t, "Test Account", client.authStore.Accounts["123"].AccountName)
	})
	
	t.Run("corrupted auth file", func(t *testing.T) {
		// Ensure directory exists
		dir := filepath.Dir(client.storePath)
		err := os.MkdirAll(dir, 0700)
		require.NoError(t, err)
		
		// Create corrupted file
		err = os.WriteFile(client.storePath, []byte("invalid json"), 0600)
		require.NoError(t, err)
		
		// Should initialize empty store
		client.loadAuthStore()
		assert.NotNil(t, client.authStore)
		assert.Empty(t, client.authStore.Accounts)
	})
}

func TestClient_saveAuthStore(t *testing.T) {
	client := createTestClient(t)
	
	// Set up test data
	client.authStore = &AuthStore{
		DefaultAccount: "123",
		Accounts: map[string]AccountToken{
			"123": {
				AccountID:   "123",
				AccountName: "Test Account",
				AccessToken: "test-token",
			},
		},
	}
	
	// Save
	err := client.saveAuthStore()
	require.NoError(t, err)
	
	// Verify file exists with correct permissions
	info, err := os.Stat(client.storePath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	
	// Verify contents
	data, err := os.ReadFile(client.storePath)
	require.NoError(t, err)
	
	var loadedStore AuthStore
	err = json.Unmarshal(data, &loadedStore)
	require.NoError(t, err)
	assert.Equal(t, "123", loadedStore.DefaultAccount)
	assert.Equal(t, "Test Account", loadedStore.Accounts["123"].AccountName)
}

func TestClient_generateState(t *testing.T) {
	client := &Client{}
	
	state := client.generateState()
	assert.NotEmpty(t, state)
	assert.True(t, len(state) >= 20) // Base64 encoded 16 bytes should be at least 20 chars
	
	// Generate multiple states to ensure they're unique
	states := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s := client.generateState()
		assert.False(t, states[s], "State should be unique")
		states[s] = true
	}
}

func TestClient_refreshToken(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		expectedError bool
		checkToken    func(t *testing.T, token *AccountToken)
	}{
		{
			name: "successful refresh",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
				
				// Parse form data
				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
				assert.Equal(t, "old-refresh-token", r.Form.Get("refresh_token"))
				
				// Return new tokens
				response := map[string]interface{}{
					"access_token":  "new-access-token",
					"refresh_token": "new-refresh-token",
					"token_type":    "Bearer",
					"expires_in":    7200,
				}
				json.NewEncoder(w).Encode(response)
			},
			expectedError: false,
			checkToken: func(t *testing.T, token *AccountToken) {
				assert.Equal(t, "new-access-token", token.AccessToken)
				assert.Equal(t, "new-refresh-token", token.RefreshToken)
				assert.Equal(t, 7200, token.ExpiresIn)
			},
		},
		{
			name: "server error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server error"))
			},
			expectedError: true,
		},
		{
			name: "invalid response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("invalid json"))
			},
			expectedError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()
			
			client := createTestClient(t)
			client.config.Endpoint.TokenURL = server.URL + "/token"
			
			token := &AccountToken{
				AccountID:    "123",
				RefreshToken: "old-refresh-token",
			}
			
			updatedToken, err := client.refreshToken(token)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkToken != nil {
					tt.checkToken(t, updatedToken)
				}
			}
		})
	}
}

func TestOAuth2Config(t *testing.T) {
	client := NewClient("test-id", "test-secret")
	
	// Verify OAuth2 configuration
	assert.Equal(t, "test-id", client.config.ClientID)
	assert.Equal(t, "test-secret", client.config.ClientSecret)
	assert.Equal(t, authURL, client.config.Endpoint.AuthURL)
	assert.Equal(t, tokenURL, client.config.Endpoint.TokenURL)
	assert.Equal(t, redirectURL, client.config.RedirectURL)
}

// Test helper to create a client with test configuration
func createTestClient(t *testing.T) *Client {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "auth.json")
	
	client := &Client{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		storePath:    storePath,
		config: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
			RedirectURL: redirectURL,
		},
		authStore: &AuthStore{
			Accounts: make(map[string]AccountToken),
		},
	}
	
	return client
}