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

	page := 1
	hasNextPage := true

	for hasNextPage {
		// Wait for rate limit
		pr.rateLimiter.Wait()

		// Prepare URL with pagination
		paginatedPath := path
		if strings.Contains(path, "?") {
			paginatedPath = fmt.Sprintf("%s&page=%d", path, page)
		} else {
			paginatedPath = fmt.Sprintf("%s?page=%d", path, page)
		}

		// Make the request
		resp, err := pr.client.doRequest("GET", paginatedPath, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		// Create a new slice to decode this page's results
		pageResults := reflect.New(sliceType)
		if err := json.NewDecoder(resp.Body).Decode(pageResults.Interface()); err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to decode page %d: %w", page, err)
		}
		resp.Body.Close()

		// Append results to the main slice
		pageSlice := pageResults.Elem()
		for i := 0; i < pageSlice.Len(); i++ {
			sliceValue.Set(reflect.Append(sliceValue, pageSlice.Index(i)))
		}

		// Check for next page in Link header
		linkHeader := resp.Header.Get("Link")
		hasNextPage = strings.Contains(linkHeader, `rel="next"`)

		// If no results on this page, we're done
		if pageSlice.Len() == 0 {
			break
		}

		page++

		// Small delay between requests to be respectful
		if hasNextPage {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// GetPage fetches a single page of results
func (pr *PaginatedRequest) GetPage(path string, page int, result interface{}) error {
	// Wait for rate limit
	pr.rateLimiter.Wait()

	// Prepare URL with pagination
	paginatedPath := path
	if strings.Contains(path, "?") {
		paginatedPath = fmt.Sprintf("%s&page=%d", path, page)
	} else {
		paginatedPath = fmt.Sprintf("%s?page=%d", path, page)
	}

	return pr.client.Get(paginatedPath, result)
}
