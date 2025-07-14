package utils

import (
	"os"
	"testing"
)

func TestPagerOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    *PagerOptions
		content string
	}{
		{
			name: "NoPager option",
			opts: &PagerOptions{
				NoPager: true,
			},
			content: "Test content",
		},
		{
			name: "Custom pager",
			opts: &PagerOptions{
				Pager:   "cat",
				Force:   true,
			},
			content: "Test content with custom pager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test just ensures the function doesn't panic
			// Actual pager behavior is hard to test in unit tests
			err := ShowInPager(tt.content, tt.opts)
			if err != nil {
				t.Errorf("ShowInPager() error = %v", err)
			}
		})
	}
}

func TestIsTerminal(t *testing.T) {
	// Test with stdout (may or may not be a terminal depending on test environment)
	_ = isTerminal(os.Stdout)
	
	// Test with nil
	if isTerminal(nil) {
		t.Error("isTerminal(nil) should return false")
	}
}