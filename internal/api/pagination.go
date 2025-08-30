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
			if currentEntry != nil && currentEntry.URL != "" {
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
