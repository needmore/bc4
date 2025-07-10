package tableprinter

import (
	"fmt"
	"time"
)

// formatRelativeTime formats time in a human-readable relative format
// following GitHub CLI's approach
func formatRelativeTime(now, timestamp time.Time) string {
	duration := now.Sub(timestamp)

	// Handle future timestamps
	if duration < 0 {
		duration = -duration
		return "in " + formatDuration(duration)
	}

	return formatDuration(duration) + " ago"
}

// formatDuration formats a duration in human-readable form
func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	minutes := int(d.Minutes())
	hours := int(d.Hours())
	days := int(d.Hours() / 24)
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case years > 0:
		if years == 1 {
			return "1 year"
		}
		return fmt.Sprintf("%d years", years)
	case months > 0:
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	case weeks > 0:
		if weeks == 1 {
			return "1 week"
		}
		return fmt.Sprintf("%d weeks", weeks)
	case days > 0:
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	case hours > 0:
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	case minutes > 0:
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	case seconds > 5:
		return fmt.Sprintf("%d seconds", seconds)
	default:
		return "just now"
	}
}
