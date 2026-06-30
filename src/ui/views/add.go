package views

import (
	"fmt"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/ui/components"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateSearching state = iota
	stateSelecting
	stateDownloading
)

type AddModel struct {
	state    state
	spinner  spinner.Model
	table    table.Model
	progress progress.Model
	query    string
}

func NewAddModel(query string) AddModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return AddModel{
		state:    stateSearching,
		spinner:  s,
		progress: components.NewProgressBar(),
		query:    query,
	}
}

func (m AddModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m AddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	switch m.state {
	case stateSelecting:
		m.table, cmd = m.table.Update(msg)
		return m, cmd

	case stateDownloading:
		newModel, newCmd := m.progress.Update(msg)
		m.progress = newModel.(progress.Model)
		return m, newCmd
	}
	return m, nil
}

func (m AddModel) View() string {
	switch m.state {
	case stateSearching:
		return fmt.Sprintf("\n %s Searching for %s...", m.spinner.View(), m.query)
	case stateSelecting:
		return "\nSelect a dependency:\n" + m.table.View()
	case stateDownloading:
		return "\nDownloading dependency:\n" + m.progress.View()
	default:
		return "Done!"
	}
}