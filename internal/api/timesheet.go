package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// TimesheetEntry represents a timesheet entry from the API
type TimesheetEntry struct {
	ID          int64     `json:"id"`
	Status      string    `json:"status"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Date        string    `json:"date"`
	Description string    `json:"description"`
	Hours       float64   `json:"hours"`
	Creator     Person    `json:"creator"`
	Parent      struct {
		ID     int64  `json:"id"`
		Title  string `json:"title"`
		Type   string `json:"type"`
		URL    string `json:"url"`
		AppURL string `json:"app_url"`
	} `json:"parent"`
	Bucket struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"bucket"`
	URL         string `json:"url"`
	AppURL      string `json:"app_url"`
	BookmarkURL string `json:"bookmark_url"`
}

// TimesheetReportOptions represents optional filters for the timesheet report
type TimesheetReportOptions struct {
	StartDate *time.Time // ISO 8601 date
	EndDate   *time.Time // ISO 8601 date
	PersonID  *int64     // Filter by person
	BucketID  *string    // Filter by project (bucket)
}

// GetTimesheetReport retrieves all timesheet entries across the account
// GET /reports/timesheet.json
// Note: This endpoint is not paginated according to the Basecamp API docs
func (c *Client) GetTimesheetReport(ctx context.Context, opts *TimesheetReportOptions) ([]TimesheetEntry, error) {
	path := "/reports/timesheet.json"

	// Build query parameters
	if opts != nil {
		params := url.Values{}
		if opts.StartDate != nil {
			params.Add("start_date", opts.StartDate.Format("2006-01-02"))
		}
		if opts.EndDate != nil {
			params.Add("end_date", opts.EndDate.Format("2006-01-02"))
		}
		if opts.PersonID != nil {
			params.Add("person_id", strconv.FormatInt(*opts.PersonID, 10))
		}
		if opts.BucketID != nil {
			params.Add("bucket_id", *opts.BucketID)
		}
		if len(params) > 0 {
			path = path + "?" + params.Encode()
		}
	}

	var entries []TimesheetEntry
	if err := c.Get(path, &entries); err != nil {
		return nil, fmt.Errorf("failed to fetch timesheet report: %w", err)
	}

	return entries, nil
}

// GetProjectTimesheet retrieves timesheet entries for a specific project
// GET /projects/{id}/timesheet.json
// Note: This endpoint supports pagination
func (c *Client) GetProjectTimesheet(ctx context.Context, projectID string) ([]TimesheetEntry, error) {
	path := fmt.Sprintf("/buckets/%s/timesheet.json", projectID)

	var entries []TimesheetEntry
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &entries); err != nil {
		return nil, fmt.Errorf("failed to fetch project timesheet: %w", err)
	}

	return entries, nil
}

// GetRecordingTimesheet retrieves timesheet entries for a specific recording within a project
// GET /projects/{id}/recordings/{id}/timesheet.json
// Note: This endpoint supports pagination
func (c *Client) GetRecordingTimesheet(ctx context.Context, projectID string, recordingID int64) ([]TimesheetEntry, error) {
	path := fmt.Sprintf("/buckets/%s/recordings/%d/timesheet.json", projectID, recordingID)

	var entries []TimesheetEntry
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &entries); err != nil {
		return nil, fmt.Errorf("failed to fetch recording timesheet: %w", err)
	}

	return entries, nil
}
