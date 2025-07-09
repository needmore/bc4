package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Common styles used across the application
var (
	// Table styles
	BaseTableStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Title styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	// Label/Value styles for detail views
	LabelStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("240"))

	ValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Help text
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Success/Error styles
	SuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42"))

	ErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	// Selection styles
	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Default indicator style (subtle)
	DefaultIndicatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42"))  // Green color for checkmark
	
	// Default row highlight style
	DefaultRowStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("228"))  // Subtle yellow background
)

// Table header style configuration
func DefaultTableHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
}

// Table selected row style configuration
func DefaultTableSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
}