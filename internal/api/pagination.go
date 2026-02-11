package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// PaginatedRequest handles paginated requests to the Basecamp API
type PaginatedRequest struct {
	client      *Client
	rateLimiter *RateLimiter
	ctx         context.Context      // nil = context.Background()
	maxPages    int                  // 0 = no limit
	pageCheck   func(page any) bool // called after each page; return false to stop pagination
}

// NewPaginatedRequest creates a new paginated request handler
func NewPaginatedRequest(client *Client) *PaginatedRequest {
	return &PaginatedRequest{
		client:      client,
		rateLimiter: GetRateLimiter(),
	}
}

// WithMaxPages sets the maximum number of pages to fetch (0 = unlimited)
func (pr *PaginatedRequest) WithMaxPages(n int) *PaginatedRequest {
	pr.maxPages = n
	return pr
}

// WithContext sets the context for all HTTP requests made during pagination.
// When the context is cancelled, in-flight requests are aborted.
func (pr *PaginatedRequest) WithContext(ctx context.Context) *PaginatedRequest {
	pr.ctx = ctx
	return pr
}

// WithPageCheck sets a callback invoked after each page is decoded.
// The callback receives the decoded page slice (e.g. []Recording) as any.
// Return false to stop pagination after the current page.
func (pr *PaginatedRequest) WithPageCheck(fn func(page any) bool) *PaginatedRequest {
	pr.pageCheck = fn
	return pr
}

// GetAll fetches all pages of results from a paginated endpoint
// The result parameter must be a pointer to a slice
func (pr *PaginatedRequest) GetAll(path string, result any) error {
	// Validate that result is a pointer to a slice
	resultType := reflect.TypeOf(result)
	if resultType.Kind() != reflect.Ptr || resultType.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to a slice")
	}

	// Get the slice value and type
	sliceValue := reflect.ValueOf(result).Elem()
	sliceType := sliceValue.Type()

	ctx := pr.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	currentPath := path
	totalFetched := 0
	pageCount := 0

	for currentPath != "" {
		// Check for context cancellation before making a request
		if err := ctx.Err(); err != nil {
			return err
		}

		// Wait for rate limit
		pr.rateLimiter.Wait()

		// Make the request with context
		resp, err := pr.client.doRequestContext(ctx, "GET", currentPath, nil)
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
		pageCount++

		// If no results on this page, we're done (safety check)
		if pageSlice.Len() == 0 {
			break
		}

		// Check max pages limit
		if pr.maxPages > 0 && pageCount >= pr.maxPages {
			break
		}

		// Check page callback — return false to stop
		if pr.pageCheck != nil && !pr.pageCheck(pageSlice.Interface()) {
			break
		}

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

		// Small delay between requests to be respectful
		if currentPath != "" {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// parseNextLinkURL extracts the next page URL from a Link header according to RFC5988
// Example: <https://3.basecampapi.com/999999999/buckets/2085958496/messages.json?page=4>; rel="next"
// Handles complex cases with quoted parameters and multiple links properly
func parseNextLinkURL(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	// Parse Link header entries more robustly
	links := parseLinkHeaderEntries(linkHeader)

	for _, link := range links {
		// Check if this link has rel="next"
		if link.hasRelation("next") {
			return link.URL
		}
	}

	return ""
}

// LinkEntry represents a single entry in a Link header
type LinkEntry struct {
	URL    string
	Params map[string]string
}

// hasRelation checks if the link has the specified relation type
func (le *LinkEntry) hasRelation(rel string) bool {
	if relValue, exists := le.Params["rel"]; exists {
		// Handle both quoted and unquoted rel values, and space-separated multiple rels
		relValue = strings.Trim(relValue, `"`)
		relations := strings.Fields(relValue)
		for _, r := range relations {
			if strings.EqualFold(r, rel) {
				return true
			}
		}
	}
	return false
}

// parseLinkHeaderEntries parses a Link header into individual entries
// This handles RFC5988 compliant parsing including quoted parameters with commas
func parseLinkHeaderEntries(linkHeader string) []LinkEntry {
	var entries []LinkEntry

	// State machine for parsing
	var currentEntry *LinkEntry
	i := 0

	for i < len(linkHeader) {
		// Skip whitespace
		for i < len(linkHeader) && (linkHeader[i] == ' ' || linkHeader[i] == '\t') {
			i++
		}
		if i >= len(linkHeader) {
			break
		}

		// Look for start of URL in angle brackets
		if linkHeader[i] == '<' {
			// Start new entry
			currentEntry = &LinkEntry{Params: make(map[string]string)}
			i++ // skip '<'

			// Find end of URL
			urlStart := i
			for i < len(linkHeader) && linkHeader[i] != '>' {
				i++
			}
			if i < len(linkHeader) {
				currentEntry.URL = linkHeader[urlStart:i]
				i++ // skip '>'
			}

			// Parse parameters after the URL
			for i < len(linkHeader) {
				// Skip whitespace and semicolons
				for i < len(linkHeader) && (linkHeader[i] == ' ' || linkHeader[i] == '\t' || linkHeader[i] == ';') {
					i++
				}
				if i >= len(linkHeader) {
					break
				}

				// Check if we hit a comma (next link) or end
				if linkHeader[i] == ',' {
					i++ // skip comma
					break
				}

				// Parse parameter name
				paramStart := i
				for i < len(linkHeader) && linkHeader[i] != '=' && linkHeader[i] != ',' && linkHeader[i] != ' ' && linkHeader[i] != '\t' {
					i++
				}
				if i > paramStart {
					paramName := linkHeader[paramStart:i]

					// Skip whitespace around =
					for i < len(linkHeader) && (linkHeader[i] == ' ' || linkHeader[i] == '\t') {
						i++
					}

					var paramValue string
					if i < len(linkHeader) && linkHeader[i] == '=' {
						i++ // skip '='

						// Skip whitespace after =
						for i < len(linkHeader) && (linkHeader[i] == ' ' || linkHeader[i] == '\t') {
							i++
						}

						if i < len(linkHeader) {
							if linkHeader[i] == '"' {
								// Quoted value - handle escapes
								i++ // skip opening quote
								valueStart := i
								for i < len(linkHeader) {
									if linkHeader[i] == '"' {
										// Check if it's escaped
										if i == 0 || linkHeader[i-1] != '\\' {
											paramValue = linkHeader[valueStart:i]
											i++ // skip closing quote
											break
										}
									}
									i++
								}
							} else {
								// Unquoted value - read until space, semicolon, or comma
								valueStart := i
								for i < len(linkHeader) && linkHeader[i] != ' ' && linkHeader[i] != '\t' &&
									linkHeader[i] != ';' && linkHeader[i] != ',' {
									i++
								}
								paramValue = linkHeader[valueStart:i]
							}
						}
					}

					currentEntry.Params[paramName] = paramValue
				}
			}

			// Add completed entry
			if currentEntry.URL != "" {
				entries = append(entries, *currentEntry)
			}
		} else {
			// Skip unexpected characters
			i++
		}
	}

	return entries
}

// extractPathFromURL converts an absolute Basecamp API URL to a relative path
// Example: https://3.basecampapi.com/999999999/buckets/123/todos.json?page=2 -> /buckets/123/todos.json?page=2
func extractPathFromURL(absoluteURL string) string {
	// If it's already a relative path, return as-is
	if strings.HasPrefix(absoluteURL, "/") {
		return absoluteURL
	}

	// Parse the URL properly using the standard library
	parsedURL, err := url.Parse(absoluteURL)
	if err != nil {
		// If URL parsing fails, return empty string to stop pagination
		return ""
	}

	// If it doesn't look like a proper URL (no scheme and not starting with /),
	// treat it as malformed
	if parsedURL.Scheme == "" && !strings.HasPrefix(absoluteURL, "/") {
		return ""
	}

	// For Basecamp API URLs, we need to extract the path after the account ID
	// Format: https://3.basecampapi.com/ACCOUNT_ID/resource...
	// or: https://3.basecampapi.com/ACCOUNT_ID/buckets/...
	path := parsedURL.Path

	// Check if this looks like a Basecamp API URL
	if parsedURL.Host != "" && strings.Contains(parsedURL.Host, "basecampapi.com") {
		// Split the path: /ACCOUNT_ID/resource... or /ACCOUNT_ID/buckets/...
		pathParts := strings.Split(strings.TrimPrefix(path, "/"), "/")

		// Need at least account_id + resource (e.g., /5624304/projects.json)
		if len(pathParts) >= 2 {
			// Check if the first part looks like an account ID (all numeric)
			// If so, skip it. Otherwise keep the full path for non-standard URLs.
			firstPart := pathParts[0]
			isAccountID := true
			for _, r := range firstPart {
				if r < '0' || r > '9' {
					isAccountID = false
					break
				}
			}

			var relativePath string
			if isAccountID && firstPart != "" {
				// Skip the account ID (first element) and reconstruct path from the rest
				relativePath = "/" + strings.Join(pathParts[1:], "/")
			} else {
				// Keep the full path for non-standard URLs
				relativePath = path
			}

			// Add query parameters if present
			if parsedURL.RawQuery != "" {
				relativePath += "?" + parsedURL.RawQuery
			}

			return relativePath
		}
	}

	// For non-Basecamp URLs or unexpected formats, return the full path + query
	// This provides better fallback behavior
	fullPath := path
	if parsedURL.RawQuery != "" {
		fullPath += "?" + parsedURL.RawQuery
	}

	return fullPath
}

// GetPage fetches a single page of results
// Note: For new code, prefer using GetAll() which handles pagination automatically.
// This method is kept for backwards compatibility and specific use cases.
func (pr *PaginatedRequest) GetPage(path string, page int, result any) error {
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
