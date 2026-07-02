package components

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	lipglossTable "charm.land/lipgloss/v2/table"
)

func NewDependencyTable(columns []table.Column, rows []table.Row) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)),
		table.WithWidth(80),
	)

	neonPink := lipgloss.Color("87")

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(neonPink).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("255"))

	s.Selected = s.Selected.
		Foreground(neonPink).
		Background(lipgloss.Color("235")). 
		Bold(true)

	s.Cell = s.Cell.
		Foreground(lipgloss.Color("255"))

	t.SetStyles(s)
	
	return t
}

func HelpTable() string {
	rows := [][]string{
		{"Command", "Description"},
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

	t := lipglossTable.New().
		Border(lipgloss.RoundedBorder()). 
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("87"))). 
		Headers(rows[0]...).
		Rows(rows[1:]...)

	t.StyleFunc(func(row, col int) lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). 
			Padding(0, 1)
	})

	return t.Render()
}