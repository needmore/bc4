package version

import (
	"strings"
	"testing"
)

func TestUserAgent(t *testing.T) {
	userAgent := UserAgent()

	// Should contain bc4-cli
	if !strings.Contains(userAgent, "bc4-cli") {
		t.Errorf("UserAgent should contain 'bc4-cli', got: %s", userAgent)
	}

	// Should contain github.com/needmore/bc4
	if !strings.Contains(userAgent, "github.com/needmore/bc4") {
		t.Errorf("UserAgent should contain 'github.com/needmore/bc4', got: %s", userAgent)
	}

	// Should follow format: bc4-cli/VERSION (github.com/needmore/bc4)
	if !strings.HasPrefix(userAgent, "bc4-cli/") {
		t.Errorf("UserAgent should start with 'bc4-cli/', got: %s", userAgent)
	}

	if !strings.HasSuffix(userAgent, "(github.com/needmore/bc4)") {
		t.Errorf("UserAgent should end with '(github.com/needmore/bc4)', got: %s", userAgent)
	}
}

func TestGet(t *testing.T) {
	info := Get()

	// Version should not be empty
	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	// Platform should contain OS and arch
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Platform should contain '/', got: %s", info.Platform)
	}

	// GoVersion should start with "go"
	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("GoVersion should start with 'go', got: %s", info.GoVersion)
	}
}