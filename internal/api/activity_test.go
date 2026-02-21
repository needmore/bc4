package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortRecordings(t *testing.T) {
	now := time.Now()

	t.Run("sorts by updated_at descending", func(t *testing.T) {
		recs := []Recording{
			{ID: 1, UpdatedAt: now.Add(-3 * time.Hour)},
			{ID: 2, UpdatedAt: now.Add(-1 * time.Hour)},
			{ID: 3, UpdatedAt: now.Add(-2 * time.Hour)},
		}
		sortRecordings(recs)

		assert.Equal(t, int64(2), recs[0].ID, "most recent first")
		assert.Equal(t, int64(3), recs[1].ID)
		assert.Equal(t, int64(1), recs[2].ID, "oldest last")
	})

	t.Run("stable sort preserves order for equal timestamps", func(t *testing.T) {
		ts := now.Add(-1 * time.Hour)
		recs := []Recording{
			{ID: 1, Title: "first", UpdatedAt: ts},
			{ID: 2, Title: "second", UpdatedAt: ts},
			{ID: 3, Title: "third", UpdatedAt: ts},
		}
		sortRecordings(recs)

		// Stable sort should preserve original order for equal keys
		assert.Equal(t, int64(1), recs[0].ID)
		assert.Equal(t, int64(2), recs[1].ID)
		assert.Equal(t, int64(3), recs[2].ID)
	})

	t.Run("empty slice", func(t *testing.T) {
		var recs []Recording
		sortRecordings(recs) // should not panic
		assert.Empty(t, recs)
	})

	t.Run("single item", func(t *testing.T) {
		recs := []Recording{{ID: 1, UpdatedAt: now}}
		sortRecordings(recs)
		assert.Equal(t, int64(1), recs[0].ID)
	})
}

func TestFilterRecordings(t *testing.T) {
	now := time.Now()
	since := now.Add(-2 * time.Hour)

	recs := []Recording{
		{ID: 1, UpdatedAt: now.Add(-1 * time.Hour), Creator: Person{ID: 100}},
		{ID: 2, UpdatedAt: now.Add(-3 * time.Hour), Creator: Person{ID: 200}},
		{ID: 3, UpdatedAt: now.Add(-30 * time.Minute), Creator: Person{ID: 100}},
		{ID: 4, UpdatedAt: now.Add(-90 * time.Minute), Creator: Person{ID: 300}},
	}

	t.Run("filter by since", func(t *testing.T) {
		opts := &ActivityListOptions{Since: &since}
		result := filterRecordings(recs, opts)

		assert.Len(t, result, 3, "should exclude record older than since")
		for _, r := range result {
			assert.NotEqual(t, int64(2), r.ID, "ID 2 is 3h old, should be excluded")
		}
	})

	t.Run("filter by person", func(t *testing.T) {
		opts := &ActivityListOptions{PersonID: 100}
		result := filterRecordings(recs, opts)

		assert.Len(t, result, 2)
		for _, r := range result {
			assert.Equal(t, int64(100), r.Creator.ID)
		}
	})

	t.Run("filter with limit", func(t *testing.T) {
		opts := &ActivityListOptions{Limit: 2}
		result := filterRecordings(recs, opts)
		assert.Len(t, result, 2)
	})

	t.Run("combined filters", func(t *testing.T) {
		opts := &ActivityListOptions{Since: &since, PersonID: 100, Limit: 1}
		result := filterRecordings(recs, opts)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(100), result[0].Creator.ID)
	})
}

// recordingsServer creates a test server that serves paginated recordings.
// It handles the /projects/recordings.json endpoint with type filtering,
// and returns recordings sorted by updated_at desc.
func recordingsServer(t *testing.T, recordingsByType map[string][]Recording) *httptest.Server {
	t.Helper()

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle project endpoint (GetProject is called first by some paths)
		if strings.Contains(r.URL.Path, "/projects/") && !strings.Contains(r.URL.Path, "recordings") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": 1, "name": "Test"}`))
			return
		}

		// Handle recordings endpoint
		recType := r.URL.Query().Get("type")
		recs, ok := recordingsByType[recType]
		if !ok {
			recs = []Recording{}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(recs)
	}))

	return srv
}

func TestListRecordings_ParallelFetch(t *testing.T) {
	now := time.Now()

	byType := map[string][]Recording{
		"Todo":     {{ID: 1, Type: "Todo", UpdatedAt: now.Add(-1 * time.Hour)}},
		"Message":  {{ID: 2, Type: "Message", UpdatedAt: now.Add(-2 * time.Hour)}},
		"Document": {{ID: 3, Type: "Document", UpdatedAt: now.Add(-30 * time.Minute)}},
		"Comment":  {{ID: 4, Type: "Comment", UpdatedAt: now.Add(-3 * time.Hour)}},
	}
	server := recordingsServer(t, byType)
	defer server.Close()

	client := &Client{
		accountID:  "123456",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	recs, err := client.ListRecordings(context.Background(), "1", nil)
	require.NoError(t, err)

	assert.Len(t, recs, 4, "should fetch all 4 types")
	// Should be sorted by updated_at desc
	assert.Equal(t, int64(3), recs[0].ID, "Document (30min ago) should be first")
	assert.Equal(t, int64(1), recs[1].ID, "Todo (1h ago) should be second")
}

func TestListRecordings_WithSinceFilter(t *testing.T) {
	now := time.Now()
	since := now.Add(-2 * time.Hour)

	byType := map[string][]Recording{
		"Todo":     {{ID: 1, Type: "Todo", UpdatedAt: now.Add(-1 * time.Hour), Creator: Person{ID: 1}}},
		"Message":  {{ID: 2, Type: "Message", UpdatedAt: now.Add(-3 * time.Hour), Creator: Person{ID: 1}}},
		"Document": {},
		"Comment":  {},
	}
	server := recordingsServer(t, byType)
	defer server.Close()

	client := &Client{
		accountID:  "123456",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	opts := &ActivityListOptions{Since: &since}
	recs, err := client.ListRecordings(context.Background(), "1", opts)
	require.NoError(t, err)

	assert.Len(t, recs, 1, "should filter out recording older than since")
	assert.Equal(t, int64(1), recs[0].ID)
}

func TestListRecordings_CustomTypes(t *testing.T) {
	now := time.Now()

	byType := map[string][]Recording{
		"Todo":    {{ID: 1, Type: "Todo", UpdatedAt: now}},
		"Message": {{ID: 2, Type: "Message", UpdatedAt: now}},
	}
	server := recordingsServer(t, byType)
	defer server.Close()

	client := &Client{
		accountID:  "123456",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	opts := &ActivityListOptions{RecordingTypes: []string{"Todo"}}
	recs, err := client.ListRecordings(context.Background(), "1", opts)
	require.NoError(t, err)

	assert.Len(t, recs, 1, "should only fetch specified type")
	assert.Equal(t, "Todo", recs[0].Type)
}

func TestListRecordings_ContextCancellation(t *testing.T) {
	// Server that delays responses
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	client := &Client{
		accountID:  "123456",
		baseURL:    srv.URL,
		httpClient: &http.Client{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.ListRecordings(ctx, "1", nil)
	assert.Error(t, err, "should fail when context is cancelled")
}

func TestListRecordingsByType_EarlyTermination(t *testing.T) {
	now := time.Now()
	cutoff := now.Add(-2 * time.Hour)

	// Page 1: recent records (should continue)
	page1 := []Recording{
		{ID: 1, UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: 2, UpdatedAt: now.Add(-90 * time.Minute)},
	}
	// Page 2: old records (should stop after this)
	page2 := []Recording{
		{ID: 3, UpdatedAt: now.Add(-3 * time.Hour)},
		{ID: 4, UpdatedAt: now.Add(-4 * time.Hour)},
	}
	// Page 3: should never be fetched
	page3 := []Recording{
		{ID: 5, UpdatedAt: now.Add(-5 * time.Hour)},
	}

	pageIndex := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages := [][]Recording{page1, page2, page3}
		if pageIndex >= len(pages) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
			return
		}

		page := pages[pageIndex]
		pageIndex++

		if pageIndex < len(pages) {
			nextURL := fmt.Sprintf("%s/123456/projects/recordings.json?page=%d", srv.URL, pageIndex+1)
			w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="next"`, nextURL))
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer srv.Close()

	client := &Client{
		accountID:  "123456",
		baseURL:    srv.URL,
		httpClient: &http.Client{},
	}

	recs, err := client.listRecordingsByType(context.Background(), "1", "Todo", &cutoff)
	require.NoError(t, err)

	// Should have page 1 (2 items) + page 2 (2 items) but NOT page 3
	// Page 1 last item (90min ago) is after cutoff (2h ago) → continue
	// Page 2 last item (4h ago) is before cutoff → stop
	assert.Len(t, recs, 4, "should fetch pages 1 and 2 but stop before page 3")
	assert.Equal(t, 2, pageIndex, "should have fetched exactly 2 pages")
}
