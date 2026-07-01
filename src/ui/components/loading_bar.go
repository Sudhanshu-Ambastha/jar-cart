package components

import (
	"github.com/charmbracelet/bubbles/progress"
)

func NewProgressBar() progress.Model {
	return progress.New(
		progress.WithGradient("#57C2FF", "#2296FF"), 
		progress.WithWidth(40),
	)
}