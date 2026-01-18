package utils

import (
	"testing"
)

func TestValidateColor(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantColor string
		wantErr   bool
	}{
		{
			name:      "valid color lowercase",
			input:     "red",
			wantColor: "red",
			wantErr:   false,
		},
		{
			name:      "valid color uppercase",
			input:     "RED",
			wantColor: "red",
			wantErr:   false,
		},
		{
			name:      "valid color mixed case",
			input:     "Red",
			wantColor: "red",
			wantErr:   false,
		},
		{
			name:      "valid color blue",
			input:     "blue",
			wantColor: "blue",
			wantErr:   false,
		},
		{
			name:      "valid color aqua",
			input:     "aqua",
			wantColor: "aqua",
			wantErr:   false,
		},
		{
			name:      "invalid color",
			input:     "magenta",
			wantColor: "",
			wantErr:   true,
		},
		{
			name:      "invalid color empty",
			input:     "",
			wantColor: "",
			wantErr:   true,
		},
		{
			name:      "all valid colors work",
			input:     "white",
			wantColor: "white",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateColor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateColor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantColor {
				t.Errorf("ValidateColor() = %v, want %v", got, tt.wantColor)
			}
		})
	}
}

func TestValidBasecampColors(t *testing.T) {
	// Ensure we have exactly 11 colors as documented in the Basecamp API
	if len(ValidBasecampColors) != 11 {
		t.Errorf("ValidBasecampColors should contain 11 colors, got %d", len(ValidBasecampColors))
	}

	// Ensure all expected colors are present
	expectedColors := map[string]bool{
		"white": true, "red": true, "orange": true, "yellow": true,
		"green": true, "blue": true, "aqua": true, "purple": true,
		"gray": true, "pink": true, "brown": true,
	}

	for _, color := range ValidBasecampColors {
		if !expectedColors[color] {
			t.Errorf("Unexpected color in ValidBasecampColors: %s", color)
		}
		delete(expectedColors, color)
	}

	if len(expectedColors) > 0 {
		t.Errorf("Missing colors in ValidBasecampColors: %v", expectedColors)
	}
}
