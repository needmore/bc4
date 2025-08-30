package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// PaginatedRequest handles paginated requests to the Basecamp API
type PaginatedRequest struct {
	client      *Client
	rateLimiter *RateLimiter
}

// NewPaginatedRequest creates a new paginated request handler
func NewPaginatedRequest(client *Client) *PaginatedRequest {
	return &PaginatedRequest{
		client:      client,
		rateLimiter: GetRateLimiter(),
	}
}

// GetAll fetches all pages of results from a paginated endpoint
// The result parameter must be a pointer to a slice
func (pr *PaginatedRequest) GetAll(path string, result interface{}) error {
	// Validate that result is a pointer to a slice
	resultType := reflect.TypeOf(result)
	if resultType.Kind() != reflect.Ptr || resultType.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to a slice")
	}

	// Get the slice value and type
	sliceValue := reflect.ValueOf(result).Elem()
	sliceType := sliceValue.Type()

	currentPath := path
	totalFetched := 0

	for currentPath != "" {
		// Wait for rate limit
		pr.rateLimiter.Wait()

		// Make the request
		resp, err := pr.client.doRequest("GET", currentPath, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch paginated results: %w", err)
		}

		// Create a new slice to decode this page's results
		pageResults := reflect.New(sliceType)
		if err := json.NewDecoder(resp.Body).Decode(pageResults.Interface()); err != nil {
			_ = resp.Body.Close()
			return fmt.Errorf("failed to decode paginated results: %w", err)
		}
		_ = resp.Body.Close()

		// Append results to the main slice
		pageSlice := pageResults.Elem()
		for i := 0; i < pageSlice.Len(); i++ {
			sliceValue.Set(reflect.Append(sliceValue, pageSlice.Index(i)))
		}

		totalFetched += pageSlice.Len()

		// Parse Link header to get next page URL according to RFC5988
		// Basecamp uses proper Link headers with rel="next"
		currentPath = ""
		linkHeader := resp.Header.Get("Link")
		if linkHeader != "" {
			nextURL := parseNextLinkURL(linkHeader)
			if nextURL != "" {
				// Convert absolute URL to relative path for our client
				currentPath = extractPathFromURL(nextURL)
			}
		}

		// If no results on this page, we're done (safety check)
		if pageSlice.Len() == 0 {
			break
		}

		// Small delay between requests to be respectful
		if currentPath != "" {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// parseNextLinkURL extracts the next page URL from a Link header
// Example: <https://3.basecampapi.com/999999999/buckets/2085958496/messages.json?page=4>; rel="next"
func parseNextLinkURL(linkHeader string) string {
	// Split by comma to handle multiple links
	links := strings.Split(linkHeader, ",")
	
	for _, link := range links {
		link = strings.TrimSpace(link)
		// Look for rel="next"
		if strings.Contains(link, `rel="next"`) {
			// Extract URL from angle brackets
			start := strings.Index(link, "<")
			end := strings.Index(link, ">")
			if start != -1 && end != -1 && start < end {
				return strings.TrimSpace(link[start+1 : end])
			}
		}
	}
	
	return ""
}

// extractPathFromURL converts an absolute Basecamp API URL to a relative path
// Example: https://3.basecampapi.com/999999999/buckets/123/todos.json?page=2 -> /buckets/123/todos.json?page=2
func extractPathFromURL(absoluteURL string) string {
	// Find the position after the account ID (third slash after https://)
	parts := strings.Split(absoluteURL, "/")
	if len(parts) >= 5 && strings.HasPrefix(absoluteURL, "https://") {
		// Reconstruct path from the bucket part onward
		pathParts := parts[4:] // Skip https:, "", domain, accountID
		return "/" + strings.Join(pathParts, "/")
	}
	
	// Fallback - if it's already a relative path, return as-is
	if strings.HasPrefix(absoluteURL, "/") {
		return absoluteURL
	}
	
	return ""
}

// GetPage fetches a single page of results
// Note: For new code, prefer using GetAll() which handles pagination automatically.
// This method is kept for backwards compatibility and specific use cases.
func (pr *PaginatedRequest) GetPage(path string, page int, result interface{}) error {
	// Wait for rate limit
	pr.rateLimiter.Wait()

	// Prepare URL with pagination
	var paginatedPath string
	if strings.Contains(path, "?") {
		paginatedPath = fmt.Sprintf("%s&page=%d", path, page)
	} else {
		paginatedPath = fmt.Sprintf("%s?page=%d", path, page)
	}

	return pr.client.Get(paginatedPath, result)
}
