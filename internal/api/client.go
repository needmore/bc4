package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/needmore/bc4/internal/config"
)

type Client struct {
	config     *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) baseURL() string {
	return fmt.Sprintf("%s/%s", c.config.BaseURL, c.config.AccountID)
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL(), path)
	
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "bc4-cli/1.0")

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

func (c *Client) Get(path string, result interface{}) error {
	resp, err := c.doRequest("GET", path, nil)
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

func (c *Client) Post(path string, payload interface{}, result interface{}) error {
	// Implementation for POST requests
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (c *Client) Put(path string, payload interface{}, result interface{}) error {
	// Implementation for PUT requests
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (c *Client) Delete(path string) error {
	// Implementation for DELETE requests
	// TODO: Implement
	return fmt.Errorf("not implemented")
}