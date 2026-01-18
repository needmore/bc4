package utils

import (
	"fmt"
	"strings"
)

// ValidBasecampColors contains all colors supported by the Basecamp API
// Reference: https://github.com/basecamp/bc3-api
var ValidBasecampColors = []string{
	"white", "red", "orange", "yellow", "green", "blue", "aqua", "purple", "gray", "pink", "brown",
}

// ValidateColor checks if the provided color is valid for Basecamp resources
// It performs case-insensitive comparison and returns the lowercase color if valid
func ValidateColor(color string) (string, error) {
	normalized := strings.ToLower(color)
	for _, validColor := range ValidBasecampColors {
		if normalized == validColor {
			return normalized, nil
		}
	}
	return "", fmt.Errorf("invalid color: %s. Available colors: %s", color, strings.Join(ValidBasecampColors, ", "))
}
