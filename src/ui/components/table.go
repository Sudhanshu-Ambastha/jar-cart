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
		{"init", "Creates an interactive project layout with JDK locking"},
		{"ls-java", "Lists all managed JDK runtimes"},
		{"cache list/ls", "Displays cached artifacts and storage usage"},
		{"cache remove/rm", "Removes cached JARs and JDKs using fuzzy matching"},
		{"cache-clear", "Clears all cached artifacts and registry data"},
		{"search <query>", "Searches Maven Central for packages"},
		{"sync", "Synchronizes dependencies and provisions project runtimes"},
		{"add <pkg>", "Adds a dependency to the project manifest"},
		{"remove <pkg>", "Removes a dependency from the project manifest"},
		{"convert <type>", "Converts manifests between supported formats (json/xml)"},
		{"run <path> [-- args...]", "Compiles and runs a project, forwarding application arguments"},
		{"run-jar <jar> [-- args...]", "Runs a JAR, forwarding application arguments"},
		{"decompile <jar>", "Decompiles JARs using Vineflower, CFR, or Procyon"},
		{"watch <path> [-- args...]", "Watches, recompiles, and restarts while preserving application arguments"},
		{"build", "Packages the project into a portable Fat JAR"},
		{"optimize <jar> <out>", "Creates a custom runtime (manifest-configured)"},
		{"self-update [version]", "Updates jar-cart or switches to a specific release"},
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
