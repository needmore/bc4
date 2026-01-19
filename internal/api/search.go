package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// SearchResult represents a search result from the Basecamp API
type SearchResult struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
	AppURL    string    `json:"app_url"`
	Creator   Person    `json:"creator"`
	Bucket    Bucket    `json:"bucket"`
	Parent    *Parent   `json:"parent,omitempty"`
}

// SearchOptions contains options for searching
type SearchOptions struct {
	Query     string   // Search query (required)
	Types     []string // Filter by recording types (Todo, Message, Document, Card)
	ProjectID string   // Scope to a specific project
	Sort      string   // Sort field: created_at or updated_at (default: updated_at)
	Direction string   // Sort direction: asc or desc (default: desc)
	Limit     int      // Maximum number of results to return (0 = no limit)
}

// ValidSearchTypes contains the valid resource types for search filtering
var ValidSearchTypes = map[string]string{
	"todo":     "Todo",
	"message":  "Message",
	"document": "Document",
	"card":     "Card",
}

// SearchOperations defines search-specific operations
type SearchOperations interface {
	Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error)
}

// Search performs a global search across all resources
func (c *Client) Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	if opts.Query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Set defaults
	if opts.Sort == "" {
		opts.Sort = "updated_at"
	}
	if opts.Direction == "" {
		opts.Direction = "desc"
	}

	// If types are specified, search each type separately and combine results
	if len(opts.Types) > 0 {
		return c.searchByTypes(ctx, opts)
	}

	// Otherwise, search without type filter (returns all types)
	return c.searchAll(ctx, opts)
}

// searchAll performs a search without type filtering
func (c *Client) searchAll(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("query", opts.Query)
	params.Set("sort", opts.Sort)
	params.Set("direction", opts.Direction)

	if opts.ProjectID != "" {
		params.Set("bucket", opts.ProjectID)
	}

	path := fmt.Sprintf("/projects/recordings.json?%s", params.Encode())

	var results []SearchResult
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &results); err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	// Apply limit if specified
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// searchByTypes performs searches for each specified type and combines results
func (c *Client) searchByTypes(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	var allResults []SearchResult

	for _, recordingType := range opts.Types {
		params := url.Values{}
		params.Set("query", opts.Query)
		params.Set("type", recordingType)
		params.Set("sort", opts.Sort)
		params.Set("direction", opts.Direction)

		if opts.ProjectID != "" {
			params.Set("bucket", opts.ProjectID)
		}

		path := fmt.Sprintf("/projects/recordings.json?%s", params.Encode())

		var typeResults []SearchResult
		pr := NewPaginatedRequest(c)
		if err := pr.GetAll(path, &typeResults); err != nil {
			return nil, fmt.Errorf("failed to search %s: %w", recordingType, err)
		}

		allResults = append(allResults, typeResults...)
	}

	// Sort combined results by the specified sort field
	sortSearchResults(allResults, opts.Sort, opts.Direction)

	// Apply limit if specified
	if opts.Limit > 0 && len(allResults) > opts.Limit {
		allResults = allResults[:opts.Limit]
	}

	return allResults, nil
}

// sortSearchResults sorts search results by the specified field and direction
func sortSearchResults(results []SearchResult, sortField, direction string) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			var shouldSwap bool

			switch sortField {
			case "created_at":
				if direction == "desc" {
					shouldSwap = results[j].CreatedAt.After(results[i].CreatedAt)
				} else {
					shouldSwap = results[j].CreatedAt.Before(results[i].CreatedAt)
				}
			default: // updated_at
				if direction == "desc" {
					shouldSwap = results[j].UpdatedAt.After(results[i].UpdatedAt)
				} else {
					shouldSwap = results[j].UpdatedAt.Before(results[i].UpdatedAt)
				}
			}

			if shouldSwap {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
