package components

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

func NewDependencyTable(columns []table.Column, rows []table.Row) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(14),
		table.WithWidth(80), 
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}

func HelpTable() table.Model {
    columns := []table.Column{
        {Title: "Command", Width: 15},
        {Title: "Description", Width: 60},
    }

    rows := []table.Row{
        {"init", "Constructs an interactive project layout with JDK locking"},
        {"ls-java", "Inventory all managed JDK runtimes"},
        {"cache list/ls", "Displays inventory and storage usage of cached artifacts"},
        {"cache remove/rm", "Fuzzy-match removal for JAR artifacts and JDKs"},
        {"cache-clear", "Wipes all cached artifacts and registry data"},
        {"search <query>", "Searches Maven Central API for packages"},
        {"sync", "Synchronizes dependencies and provisions local runtimes"},
        {"add <pkg>", "Adds an artifact dependency to your manifest"},
        {"remove <pkg>", "Removes dependency and cleans local links"},
        {"convert <type>", "Translates manifest formats (json/xml)"},
        {"run <path>", "Compiles and executes with the project-locked JDK"},
        {"run-jar <jar>", "Runs built JAR with isolated environment"},
        {"decompile <jar>", "Extracts source via (vineflower|cfr|procyon)"},
        {"watch <path>", "Reactive file-watcher with SHA256 integrity checks"},
        {"build", "Packages project into a portable Fat JAR"},
        {"help", "Displays this documentation"},
    }

    return NewDependencyTable(columns, rows)
}