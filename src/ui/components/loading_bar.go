package components

import (
	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
)

func NewProgressBar() progress.Model {
	return progress.New(
		progress.WithColors(
			lipgloss.Color("#57"),
			lipgloss.Color("#229"), 
		),
		progress.WithWidth(40),
	)
}