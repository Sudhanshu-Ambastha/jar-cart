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
		{"init", "Constructs an interactive or default blueprint layout"},
		{"cache-clear", "Clears all cached blueprints and registry data"},
		{"search <query>", "Searches Maven Central API for packages"},
		{"sync", "Synchronizes dependencies"},
		{"add <pkg>", "Appends an artifact dependency to your manifest"},
		{"remove <pkg>", "Strips an artifact marker and cleans up the local JAR"},
		{"convert <type>", "Translates configuration contexts (json|xml)"},
		{"run <path>", "Compiles and runs a Java source file or script"},
		{"run-jar <jar>", "Runs the built JAR with all dependencies"},
		{"decompile <jar>", "Extracts source code via --engine (vineflower|cfr|procyon)"},
		{"watch <path>", "Starts a reactive file-watcher for live reloads"},
		{"build", "Packages the project into a portable Fat JAR"},
		{"help", "Displays this documentation"},
	}

	return NewDependencyTable(columns, rows)
}