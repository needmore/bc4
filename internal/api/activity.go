package api

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"
)

// Event represents a Basecamp activity event
type Event struct {
	ID            int64     `json:"id"`
	Action        string    `json:"action"`
	CreatedAt     time.Time `json:"created_at"`
	RecordingType string    `json:"recording_type"`
	Recording     Recording `json:"recording"`
	Creator       Person    `json:"creator"`
	Bucket        Bucket    `json:"bucket"`
}

// Recording represents a Basecamp recording (generic content item)
type Recording struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
	AppURL    string    `json:"app_url"`
	Creator   Person    `json:"creator"`
	Bucket    Bucket    `json:"bucket"`
	Parent    *Parent   `json:"parent,omitempty"`
}

// Bucket represents a Basecamp bucket (project container)
type Bucket struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Parent represents a parent recording
type Parent struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	AppURL string `json:"app_url"`
}

// ActivityListOptions contains options for listing activity
type ActivityListOptions struct {
	Since          *time.Time // Filter events since this time
	RecordingTypes []string   // Filter by recording types (todo, message, document, etc.)
	PersonID       int64      // Filter by person ID (creator)
	Limit          int        // Maximum number of events to return
}

// ListEvents returns activity events for a recording
func (c *Client) ListEvents(ctx context.Context, projectID string, recordingID int64) ([]Event, error) {
	var events []Event
	path := fmt.Sprintf("/buckets/%s/recordings/%d/events.json", projectID, recordingID)

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &events); err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return events, nil
}

// ListRecordings returns all recordings (activity items) for a project.
// Types are fetched in parallel using errgroup — if one type fails or the
// context is cancelled, all in-flight fetches are aborted. When opts.Since
// is set, pagination stops early once records older than the cutoff are encountered.
func (c *Client) ListRecordings(ctx context.Context, projectID string, opts *ActivityListOptions) ([]Recording, error) {
	// Default types to fetch if none specified
	typesToFetch := []string{"Todo", "Message", "Document", "Comment"}

	if opts != nil && len(opts.RecordingTypes) > 0 {
		typesToFetch = opts.RecordingTypes
	}

	// Extract options for per-type fetching
	var since *time.Time
	if opts != nil {
		since = opts.Since
	}

	// Fetch all types in parallel; cancel siblings on first error
	g, gctx := errgroup.WithContext(ctx)
	results := make([][]Recording, len(typesToFetch))

	for i, recordingType := range typesToFetch {
		g.Go(func() error {
			recs, err := c.listRecordingsByType(gctx, projectID, recordingType, since)
			if err != nil {
				return fmt.Errorf("failed to list %s recordings: %w", recordingType, err)
			}
			results[i] = recs
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Collect results
	var allRecordings []Recording
	for _, recs := range results {
		allRecordings = append(allRecordings, recs...)
	}

	// Sort by updated_at descending (stable to preserve order for equal timestamps)
	sortRecordings(allRecordings)

	// Apply filtering if options provided
	if opts != nil {
		allRecordings = filterRecordings(allRecordings, opts)
	}

	return allRecordings, nil
}

// listRecordingsByType fetches recordings of a specific type for a project.
// When since is non-nil, pagination stops early once all items on a page
// are older than the cutoff (data arrives sorted by updated_at desc).
// The context is propagated to all HTTP requests for cancellation support.
func (c *Client) listRecordingsByType(ctx context.Context, projectID string, recordingType string, since *time.Time) ([]Recording, error) {
	var recordings []Recording

	// Build query params
	params := url.Values{}
	params.Set("bucket", projectID)
	params.Set("type", recordingType)
	params.Set("sort", "updated_at")
	params.Set("direction", "desc")

	path := fmt.Sprintf("/projects/recordings.json?%s", params.Encode())

	pr := NewPaginatedRequest(c).WithContext(ctx)

	// Early termination: stop paginating once the last item on a page
	// is older than our since cutoff. Since results are sorted by
	// updated_at desc, all subsequent pages will only have older records.
	if since != nil {
		cutoff := *since
		pr.WithPageCheck(func(page any) bool {
			recs, ok := page.([]Recording)
			if !ok || len(recs) == 0 {
				return false
			}
			last := recs[len(recs)-1]
			return !last.UpdatedAt.Before(cutoff)
		})
	}

	if err := pr.GetAll(path, &recordings); err != nil {
		return nil, err
	}

	return recordings, nil
}

// sortRecordings sorts recordings by updated_at in descending order.
// Uses stable sort to preserve relative order of items with equal timestamps.
func sortRecordings(recordings []Recording) {
	sort.SliceStable(recordings, func(i, j int) bool {
		return recordings[i].UpdatedAt.After(recordings[j].UpdatedAt)
	})
}

// filterRecordings applies filtering options to recordings
func filterRecordings(recordings []Recording, opts *ActivityListOptions) []Recording {
	capacity := len(recordings)
	if opts.Limit > 0 && opts.Limit < capacity {
		capacity = opts.Limit
	}
	filtered := make([]Recording, 0, capacity)

	for _, r := range recordings {
		// Filter by since time
		if opts.Since != nil && r.UpdatedAt.Before(*opts.Since) {
			continue
		}

		// Filter by person ID
		if opts.PersonID > 0 && r.Creator.ID != opts.PersonID {
			continue
		}

		filtered = append(filtered, r)

		// Apply limit
		if opts.Limit > 0 && len(filtered) >= opts.Limit {
			break
		}
	}

	return filtered
}

// GetRecording returns a specific recording
func (c *Client) GetRecording(ctx context.Context, projectID string, recordingID int64) (*Recording, error) {
	var recording Recording
	path := fmt.Sprintf("/buckets/%s/recordings/%d.json", projectID, recordingID)

	if err := c.Get(path, &recording); err != nil {
		return nil, fmt.Errorf("failed to get recording: %w", err)
	}

	return &recording, nil
}
