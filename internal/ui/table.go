package ui

import (
	"os"

	"golang.org/x/term"
)

// GetTerminalWidth returns the current terminal width with a fallback
func GetTerminalWidth() int {
	defaultWidth := 100
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width - 2 // Leave minimal margin to prevent wrapping
	}
	return defaultWidth
}

// GetTerminalHeight returns the current terminal height with a fallback
func GetTerminalHeight() int {
	defaultHeight := 24
	if _, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil && height > 0 {
		return height
	}
	return defaultHeight
}
