package utils

type MenuOption struct {
	Name  string
	Color string
}

func InteractiveInit(projectName, passedFramework, passedStrategy, passedLanguage string) (string, string, string, string) {
	framework := passedFramework
	if framework == "" {
		framework = "Vanilla Java Application"
	}

	strategy := "no-build"

	lang := passedLanguage
	if lang == "" {
		lang = "Java"
	}

	return projectName, framework, strategy, lang
}