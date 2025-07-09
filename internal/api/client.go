package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://3.basecampapi.com"
	userAgent      = "bc4-cli/1.0.0 (github.com/needmore/bc4)"
)

type Client struct {
	accountID    string
	accessToken  string
	httpClient   *http.Client
	baseURL      string
}

func NewClient(accountID, accessToken string) *Client {
	return &Client{
		accountID:   accountID,
		accessToken: accessToken,
		baseURL:     defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) getBaseURL() string {
	return fmt.Sprintf("%s/%s", c.baseURL, c.accountID)
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
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

// Project represents a Basecamp project
type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// GetProjects fetches all projects for the account
func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	var projects []Project
	
	// Basecamp API returns projects at /projects.json
	if err := c.Get("/projects.json", &projects); err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}
	
	return projects, nil
}