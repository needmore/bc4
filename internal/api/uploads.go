package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/needmore/bc4/internal/errors"
	"github.com/needmore/bc4/internal/version"
)

// Upload represents a Basecamp upload (file attachment)
type Upload struct {
	ID             int64     `json:"id"`
	Filename       string    `json:"filename"`
	ContentType    string    `json:"content_type"`
	ByteSize       int64     `json:"byte_size"`
	DownloadURL    string    `json:"download_url"`
	AppDownloadURL string    `json:"app_download_url"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetUpload fetches upload details by ID
func (c *Client) GetUpload(ctx context.Context, bucketID string, uploadID int64) (*Upload, error) {
	path := fmt.Sprintf("/buckets/%s/uploads/%d.json", bucketID, uploadID)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var upload Upload
	if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &upload, nil
}

// DownloadAttachment downloads a file from a download URL to the specified path
func (c *Client) DownloadAttachment(ctx context.Context, downloadURL, destPath string) error {
	// Create the request with OAuth authentication
	req, err := c.createAuthenticatedRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download attachment: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		switch resp.StatusCode {
		case 403:
			return errors.NewAPIError(resp.StatusCode, fmt.Sprintf("permission denied: %s", string(body)), nil)
		case 404:
			return errors.NewNotFoundError("attachment", "", fmt.Errorf("attachment not found: %s", string(body)))
		default:
			return errors.NewAPIError(resp.StatusCode, string(body), nil)
		}
	}

	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create the destination file
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	// Copy the response body to the file
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write attachment to file: %w", err)
	}

	// Set file permissions to 0644
	if err := outFile.Chmod(0644); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// createAuthenticatedRequest creates an HTTP request with OAuth authentication
func (c *Client) createAuthenticatedRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("User-Agent", version.UserAgent())

	return req, nil
}

// ExtractUploadID extracts the upload ID from various URL formats
// Supports:
// - https://3.basecampapi.com/{account}/buckets/{bucket}/uploads/{id}/download/{filename}
// - https://3.basecamp.com/{account}/buckets/{bucket}/uploads/{id}
func ExtractUploadID(url string) (int64, error) {
	// Pattern for API download URL
	apiPattern := regexp.MustCompile(`/uploads/(\d+)/download`)
	if matches := apiPattern.FindStringSubmatch(url); len(matches) > 1 {
		var id int64
		if _, err := fmt.Sscanf(matches[1], "%d", &id); err != nil {
			return 0, fmt.Errorf("failed to parse upload ID: %w", err)
		}
		return id, nil
	}

	// Pattern for app URL
	appPattern := regexp.MustCompile(`/uploads/(\d+)`)
	if matches := appPattern.FindStringSubmatch(url); len(matches) > 1 {
		var id int64
		if _, err := fmt.Sscanf(matches[1], "%d", &id); err != nil {
			return 0, fmt.Errorf("failed to parse upload ID: %w", err)
		}
		return id, nil
	}

	return 0, fmt.Errorf("could not extract upload ID from URL: %s", url)
}
