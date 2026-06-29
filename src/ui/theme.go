package ui

import "github.com/charmbracelet/lipgloss"

var (
	PrimaryColor   = lipgloss.Color("205") // A nice vibrant blue/purple
	SuccessColor   = lipgloss.Color("42")  // Vibrant green
	ErrorColor     = lipgloss.Color("196") // Bright red
	HighlightColor = lipgloss.Color("214") // Orange for focus

	TitleStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)
)