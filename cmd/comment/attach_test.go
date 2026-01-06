package comment

import (
	"testing"
	"time"

	"github.com/needmore/bc4/internal/api"
)

func TestNewestCommentID(t *testing.T) {
	comments := []api.Comment{
		{ID: 1, CreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
		{ID: 2, CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 3, CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}, // tie on time, higher ID wins
	}

	got := newestCommentID(comments)
	if got != 3 {
		t.Fatalf("expected newest comment ID 3, got %d", got)
	}
}
