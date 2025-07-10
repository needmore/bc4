package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// OutputFormat represents the desired output format
type OutputFormat string

const (
	// OutputFormatTable renders as a human-readable table
	OutputFormatTable OutputFormat = "table"
	// OutputFormatJSON renders as JSON
	OutputFormatJSON OutputFormat = "json"
	// OutputFormatTSV renders as tab-separated values
	OutputFormatTSV OutputFormat = "tsv"
)

// ParseOutputFormat parses a string into an OutputFormat
func ParseOutputFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(s) {
	case "table", "":
		return OutputFormatTable, nil
	case "json":
		return OutputFormatJSON, nil
	case "tsv":
		return OutputFormatTSV, nil
	default:
		return "", fmt.Errorf("unknown output format: %s", s)
	}
}

// IsTerminal returns true if the given writer is a terminal
func IsTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		// Use the same check as in table.go
		fd := int(f.Fd())
		return isatty(fd)
	}
	return false
}

// isatty is a helper to check if a file descriptor is a terminal
func isatty(fd int) bool {
	// Check if we're forcing TTY mode
	if os.Getenv("BC4_FORCE_TTY") != "" || os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	// Try to get terminal size - if it works, it's a terminal
	_, _, err := term.GetSize(fd)
	return err == nil
}