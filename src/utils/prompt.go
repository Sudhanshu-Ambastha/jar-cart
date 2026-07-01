package utils

import (
	"github.com/charmbracelet/huh"
)

func InteractiveInit(detectedVersion string) (string, string) {
	var manifestFormat string
	javaVersion := detectedVersion

	fields := []huh.Field{
		huh.NewSelect[string]().
			Title("Manifest Format").
			Options(huh.NewOption("JSON", "json"), huh.NewOption("XML", "xml")).
			Value(&manifestFormat),
	}

	if javaVersion == "" {
		fields = append(fields, huh.NewSelect[string]().
			Title("Java Version").
			Options(huh.NewOption("21", "21"), huh.NewOption("25", "25")).
			Value(&javaVersion),
		)
	}

	form := huh.NewForm(
		huh.NewGroup(fields...),
	)
	_ = form.Run()
	
	return manifestFormat, javaVersion
}

func InteractiveSelection(options map[string]string) (string, bool) {
	var selectedValue string
	choices := []huh.Option[string]{}

	for display, value := range options {
		choices = append(choices, huh.NewOption(display, value))
	}

	if IsOnline() {
		choices = append(choices, huh.NewOption("🌐 Search online (Maven Central)...", "ONLINE"))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a version or search online:").
				Options(choices...).
				Value(&selectedValue),
		),
	)
	_ = form.Run()
	
	if selectedValue == "ONLINE" || selectedValue == "" {
		return "", false
	}
	return selectedValue, true 
}