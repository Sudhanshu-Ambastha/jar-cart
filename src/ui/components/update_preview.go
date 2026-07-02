package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func UpdateNotification(current, latest string) string {

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D9FF")).
		Padding(1, 2).
		Width(62)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7DF9FF"))

	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C8A9A"))

	currentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B5B5B5")).
		Bold(true)

	latestStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#39FF14")).
		Bold(true)

	arrow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	command := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF"))

	body := fmt.Sprintf(
		"%s\n\n%s %s %s %s\n\n%s\n  %s",
		title.Render("🚀 Update Available"),
		label.Render("Version:"),
		currentStyle.Render(current),
		arrow.Render("→"),
		latestStyle.Render(latest),
		label.Render("Run:"),
		command.Render("❯ jar-cart self-update"),
	)

	return box.Render(body)
}