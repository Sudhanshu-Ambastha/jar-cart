package components

import (
	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
)

func NewStyledSpinner(s spinner.Spinner, color string) spinner.Model {
	sp := spinner.New()
	sp.Spinner = s
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	return sp
}