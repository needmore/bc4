package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testItem is a simple struct for pagination tests
type testItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// newTestPaginatedServer creates an httptest server that serves paginated JSON arrays.
// pages is a slice of slices — each inner slice is one page of results.
func newTestPaginatedServer(t *testing.T, pages [][]testItem) *httptest.Server {
	t.Helper()
	pageIndex := 0

	mux := http.NewServeMux()
	var srv *httptest.Server

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if pageIndex >= len(pages) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
			return
		}

		page := pages[pageIndex]
		pageIndex++

		if pageIndex < len(pages) {
			nextURL := fmt.Sprintf("%s/123456/items.json?page=%d", srv.URL, pageIndex+1)
			w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="next"`, nextURL))
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(page)
	})

	srv = httptest.NewServer(mux)
	return srv
}

func newTestClient(serverURL string) *Client {
	return &Client{
		accountID:  "123456",
		baseURL:    serverURL,
		httpClient: &http.Client{},
	}
}

func TestGetAll_SinglePage(t *testing.T) {
	pages := [][]testItem{
		{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}},
	}
	server := newTestPaginatedServer(t, pages)
	defer server.Close()

	client := newTestClient(server.URL)
	pr := NewPaginatedRequest(client)

	var items []testItem
	err := pr.GetAll("/items.json", &items)

	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, 1, items[0].ID)
	assert.Equal(t, 2, items[1].ID)
}

func TestGetAll_MultiplePages(t *testing.T) {
	pages := [][]testItem{
		{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}},
		{{ID: 3, Name: "c"}},
	}
	server := newTestPaginatedServer(t, pages)
	defer server.Close()

	client := newTestClient(server.URL)
	pr := NewPaginatedRequest(client)

	var items []testItem
	err := pr.GetAll("/items.json", &items)

	require.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, 3, items[2].ID)
}

func TestGetAll_WithMaxPages(t *testing.T) {
	pages := [][]testItem{
		{{ID: 1, Name: "a"}},
		{{ID: 2, Name: "b"}},
		{{ID: 3, Name: "c"}},
	}
	server := newTestPaginatedServer(t, pages)
	defer server.Close()

	client := newTestClient(server.URL)
	pr := NewPaginatedRequest(client).WithMaxPages(2)

	var items []testItem
	err := pr.GetAll("/items.json", &items)

	require.NoError(t, err)
	assert.Len(t, items, 2, "should stop after 2 pages")
	assert.Equal(t, 1, items[0].ID)
	assert.Equal(t, 2, items[1].ID)
}

func TestGetAll_WithPageCheck_StopsEarly(t *testing.T) {
	pages := [][]testItem{
		{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}},
		{{ID: 3, Name: "c"}, {ID: 4, Name: "d"}},
		{{ID: 5, Name: "e"}},
	}
	server := newTestPaginatedServer(t, pages)
	defer server.Close()

	client := newTestClient(server.URL)
	pr := NewPaginatedRequest(client).WithPageCheck(func(page any) bool {
		items := page.([]testItem)
		// Stop after seeing ID >= 3
		last := items[len(items)-1]
		return last.ID < 3
	})

	var items []testItem
	err := pr.GetAll("/items.json", &items)

	require.NoError(t, err)
	// First page (IDs 1,2) passes check (last.ID=2 < 3), continues
	// Second page (IDs 3,4) fails check (last.ID=4 >= 3), stops
	assert.Len(t, items, 4, "should include pages 1 and 2 but stop before page 3")
}

func TestGetAll_WithContext_CancellationStopsPagination(t *testing.T) {
	pages := [][]testItem{
		{{ID: 1, Name: "a"}},
		{{ID: 2, Name: "b"}},
		{{ID: 3, Name: "c"}},
	}
	server := newTestPaginatedServer(t, pages)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	client := newTestClient(server.URL)
	pr := NewPaginatedRequest(client).WithContext(ctx).WithPageCheck(func(page any) bool {
		// Cancel context after first page
		cancel()
		return true // would continue, but context is cancelled
	})

	var items []testItem
	err := pr.GetAll("/items.json", &items)

	// Should get first page's items, then error on context cancellation
	assert.Error(t, err)
	assert.Len(t, items, 1, "should have fetched first page before cancel")
}

func TestGetAll_EmptyResponse(t *testing.T) {
	pages := [][]testItem{
		{}, // empty page
	}
	server := newTestPaginatedServer(t, pages)
	defer server.Close()

	client := newTestClient(server.URL)
	pr := NewPaginatedRequest(client)

	var items []testItem
	err := pr.GetAll("/items.json", &items)

	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestGetAll_InvalidResultType(t *testing.T) {
	client := &Client{}
	pr := NewPaginatedRequest(client)

	var notASlice string
	err := pr.GetAll("/items.json", &notASlice)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "result must be a pointer to a slice")

	var notAPointer []testItem
	err = pr.GetAll("/items.json", notAPointer)
	assert.Error(t, err)
}
