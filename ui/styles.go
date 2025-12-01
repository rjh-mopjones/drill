package ui

import (
	"hash/fnv"

	"github.com/charmbracelet/lipgloss"
)

// Service colors - bright colors for text visibility
var serviceColors = []lipgloss.Color{
	lipgloss.Color("#42a5f5"), // Blue
	lipgloss.Color("#66bb6a"), // Green
	lipgloss.Color("#ab47bc"), // Purple
	lipgloss.Color("#ffca28"), // Amber
	lipgloss.Color("#26c6da"), // Cyan
	lipgloss.Color("#ef5350"), // Red
	lipgloss.Color("#9ccc65"), // Light green
	lipgloss.Color("#26a69a"), // Teal
	lipgloss.Color("#ff7043"), // Deep orange
	lipgloss.Color("#8d6e63"), // Brown
}

// Correlation ID colors - brighter colors for text visibility
var correlationColors = []lipgloss.Color{
	lipgloss.Color("#ff6b6b"), // Coral red
	lipgloss.Color("#4ecdc4"), // Teal
	lipgloss.Color("#ffe66d"), // Yellow
	lipgloss.Color("#95e1d3"), // Mint
	lipgloss.Color("#f38181"), // Salmon
	lipgloss.Color("#aa96da"), // Lavender
	lipgloss.Color("#fcbad3"), // Pink
	lipgloss.Color("#a8d8ea"), // Light blue
	lipgloss.Color("#dcedc1"), // Light green
	lipgloss.Color("#ffd3b6"), // Peach
}

var (
	// Base styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#5c6bc0")).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#3949ab")).
			Padding(0, 2).
			MarginBottom(1)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#e0e0e0")).
				Background(lipgloss.Color("#424242")).
				Padding(0, 1)

	FailedCommandStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff5252"))

	SuccessCommandStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#69f0ae"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5c6bc0"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginTop(1)

	SelectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#5c6bc0")).
				Foreground(lipgloss.Color("#ffffff"))
)

// GetServiceColor returns a consistent text color for a service name
func GetServiceColor(serviceName string) lipgloss.Color {
	h := fnv.New32a()
	h.Write([]byte(serviceName))
	index := int(h.Sum32()) % len(serviceColors)
	return serviceColors[index]
}

// GetCorrelationColor returns a consistent text color for a correlation ID
func GetCorrelationColor(correlationID string) lipgloss.Color {
	h := fnv.New32a()
	h.Write([]byte(correlationID))
	index := int(h.Sum32()) % len(correlationColors)
	return correlationColors[index]
}

// CreateServiceStyle returns a style with the appropriate text color for the service
func CreateServiceStyle(serviceName string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(GetServiceColor(serviceName)).
		Bold(true)
}

// CreateCorrelationStyle returns a style with the appropriate color for the correlation ID
func CreateCorrelationStyle(correlationID string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(GetCorrelationColor(correlationID)).
		Bold(true)
}
