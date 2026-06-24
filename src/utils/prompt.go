package utils

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

type MenuOption struct {
	Name  string
	Color string
}

func InteractiveInit(projectName, passedFramework, passedStrategy, passedLanguage string) {
	fmt.Println("\033[36m🛒 Welcome to jar-cart workspace construction engine!\033[0m\n")

	frameworkChoice := passedFramework
	if frameworkChoice == "" {
		frameworks := []MenuOption{
			{Name: "Vanilla Java Application", Color: "\033[33m"},
			{Name: "Spring Boot",              Color: "\033[32m"},
			{Name: "Quarkus",                  Color: "\033[36m"},
			{Name: "MicroProfile",             Color: "\033[34m"},
			{Name: "JavaFX Desktop",           Color: "\033[35m"},
			{Name: "Micronaut",                Color: "\033[31m"},
			{Name: "Graeval Development Kit",  Color: "\033[90m"},
		}

		frameworkTemplates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\033[32m❯\033[0m {{ .Color }}\033[1m{{ .Name }}\033[0m",
			Inactive: "  {{ .Color }}{{ .Name }}\033[0m",
			Selected: "\033[32m✔\033[0m Selected Framework: \033[1m{{ .Name }}\033[0m",
		}

		frameworkPrompt := promptui.Select{
			Label:     "\033[36m◆\033[0m Select a framework / layout architecture:",
			Items:     frameworks,
			Templates: frameworkTemplates,
			Size:      10,
		}

		idx, _, err := frameworkPrompt.Run()
		if err != nil {
			fmt.Println("❌ Initialization aborted.")
			return
		}
		frameworkChoice = frameworks[idx].Name
	} else {
		lowerFramework := strings.ToLower(frameworkChoice)
		if lowerFramework == "spring" || lowerFramework == "springboot" {
			frameworkChoice = "Spring Boot"
		} else if lowerFramework == "quarkus" {
			frameworkChoice = "Quarkus"
		} else if lowerFramework == "micronaut" {
			frameworkChoice = "Micronaut"
		} else if lowerFramework == "javafx" {
			frameworkChoice = "JavaFX Desktop"
		} else {
			frameworkChoice = strings.Title(lowerFramework)
		}
		fmt.Printf("\033[32m✔\033[0m Framework (via flag): \033[1m%s\033[0m\n", frameworkChoice)
	}

	strategyChoice := passedStrategy
	if strategyChoice == "" {
		strategies := []MenuOption{
			{Name: "no-build (Work with source code directly)", Color: "\033[36m"},
			{Name: "maven (Create from archetype)",            Color: "\033[34m"},
			{Name: "gradle (Modern Groovy/Kotlin)",            Color: "\033[33m"},
		}

		strategyTemplates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\033[32m❯\033[0m {{ .Color }}\033[1m{{ .Name }}\033[0m",
			Inactive: "  {{ .Color }}{{ .Name }}\033[0m",
			Selected: "\033[32m✔\033[0m Selected Strategy: \033[1m{{ .Name }}\033[0m",
		}

		strategyPrompt := promptui.Select{
			Label:     "\033[36m◆\033[0m Select execution strategy ecosystem:",
			Items:     strategies,
			Templates: strategyTemplates,
			Size:      10,
		}

		sIdx, _, err := strategyPrompt.Run()
		if err != nil {
			return
		}
		strategyChoice = strategies[sIdx].Name
	} else {
		strategyChoice = strings.ToLower(strategyChoice)
		fmt.Printf("\033[32m✔\033[0m Strategy (via flag): \033[1m%s\033[0m\n", strategyChoice)
	}

	strategyClean := strings.Split(strategyChoice, " ")[0]
	languageChoice := passedLanguage

	if strategyClean == "gradle" && languageChoice == "" {
		languages := []MenuOption{
			{Name: "Java",                      Color: "\033[34m"},
			{Name: "Kotlin (build.gradle.kts)", Color: "\033[35m"},
			{Name: "Groovy (build.gradle)",     Color: "\033[33m"},
		}

		langTemplates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\033[32m❯\033[0m {{ .Color }}\033[1m{{ .Name }}\033[0m",
			Inactive: "  {{ .Color }}{{ .Name }}\033[0m",
			Selected: "\033[32m✔\033[0m Target Engine: \033[1m{{ .Name }}\033[0m",
		}

		langPrompt := promptui.Select{
			Label:     "\033[36m◆\033[0m Select Language DSL script variant:",
			Items:     languages,
			Templates: langTemplates,
			Size:      10,
		}
		lIdx, _, err := langPrompt.Run()
		if err != nil {
			return
		}
		languageChoice = languages[lIdx].Name
	} else if strategyClean == "gradle" && languageChoice != "" {
		languageChoice = strings.Title(strings.ToLower(languageChoice))
		fmt.Printf("\033[32m✔\033[0m Language DSL (via flag): \033[1m%s\033[0m\n", languageChoice)
	} else {
		languageChoice = "Java"
	}

	fmt.Printf("\n⚡ Structuring scaffolding blocks for \033[34m%s\033[0m...\n", projectName)
	fmt.Printf("📦 Blueprint: \033[1m%s\033[0m | Strategy: \033[1m%s\033[0m | Language: \033[1m%s\033[0m\n", frameworkChoice, strategyClean, languageChoice)

	ExecuteScaffold(projectName, projectName, frameworkChoice, strategyClean, languageChoice, "25")

	fmt.Println("\n\033[32m✨ Project workspace successfully configured! Happy coding! 🛒🏎️💨\033[0m")

	LaunchWorkspace(projectName)
}