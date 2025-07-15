package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://3.basecampapi.com"
	userAgent      = "bc4-cli/1.0.0 (github.com/needmore/bc4)"
)

// BaseClient provides core HTTP functionality for all API clients
type BaseClient struct {
	accountID   string
	accessToken string
	httpClient  *http.Client
	baseURL     string
}

// NewBaseClient creates a new base client with authentication
func NewBaseClient(accountID, accessToken string) *BaseClient {
	return &BaseClient{
		accountID:   accountID,
		accessToken: accessToken,
		baseURL:     defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccountID returns the account ID
func (c *BaseClient) GetAccountID() string {
	return c.accountID
}

// getBaseURL returns the full base URL including account ID
func (c *BaseClient) getBaseURL() string {
	return fmt.Sprintf("%s/%s", c.baseURL, c.accountID)
}

// DoRequest performs an HTTP request with authentication
func (c *BaseClient) DoRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.getBaseURL(), path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return resp, nil
}

// Get performs a GET request and decodes the response
func (c *BaseClient) Get(path string, result interface{}) error {
	resp, err := c.DoRequest("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Post performs a POST request with a JSON payload
func (c *BaseClient) Post(path string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}

	resp, err := c.DoRequest("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Put performs a PUT request with a JSON payload
func (c *BaseClient) Put(path string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}

	resp, err := c.DoRequest("PUT", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Delete performs a DELETE request
func (c *BaseClient) Delete(path string) error {
	resp, err := c.DoRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}