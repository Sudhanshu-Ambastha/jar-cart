package utils

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
)

func InteractiveInit(detectedVersion string) (string, string) {
    manifestFormat := "json" 
    selectedVersion := "25"  
    var customVersion string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Manifest Format").
                Options(huh.NewOption("JSON", "json"), huh.NewOption("XML", "xml")).
                Value(&manifestFormat),

            huh.NewSelect[string]().
                Title("Java Version").
                Options(
                    huh.NewOption("21 (LTS)", "21"),
                    huh.NewOption("25 (Latest LTS)", "25"),
                    huh.NewOption("Custom Version", "custom"),
                ).
                Value(&selectedVersion),
        ),
        huh.NewGroup(
            huh.NewInput().
                Title("Enter Java Version (e.g., 17)").
                Value(&customVersion),
        ).WithHideFunc(func() bool {
            return selectedVersion != "custom"
        }),
    )
    if err := form.Run(); err != nil {
        fmt.Println("Error running form:", err)
        os.Exit(1)
    }

    if selectedVersion == "custom" {
        return manifestFormat, customVersion
    }
    return manifestFormat, selectedVersion
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