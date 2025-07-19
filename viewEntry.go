package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//for viewing entries if they're clicked. only viewable if the user clicks something (v? while not on search options) if "e" is clicked
//while on entry, entry is made edit-able by writing entry.

type viewLog struct {
	viewPort viewport.Model
	header   string
	date     string
}

func (m *model) setRow(row table.Row) {
	m.tab.eView.header = row[0]
	m.tab.eView.date = row[1]
	m.tab.eView.viewPort.SetContent(row[3])

}

func (m model) viewportInit() model {

	m.tab.eView.viewPort = viewport.New((m.width * int(m.config.Height)), (m.width * int(m.config.Height)))
	//we can assume that data is already loaded in, as this functino shouldn't even be triggered if data isn't loaded into table
	selEntry := m.tab.table.SelectedRow()

	if len(selEntry) == 0 {
		m.tab.eView.viewPort.SetContent("")
		return m
	}
	m.setRow(selEntry)
	return m
}

func (m *model) viewportUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.tab.eView.viewPort, cmd = m.tab.eView.viewPort.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) viewportView() string {

	selEntry := m.tab.table.SelectedRow()

	if selEntry != nil {

		if len(selEntry) == 0 {
			m.tab.eView.viewPort.SetContent("")
		} else {
			m.setRow(selEntry)
			m.tab.eView.viewPort.Height = lipgloss.Height(selEntry[3])
		}
		//}
		m.tab.eView.viewPort.SetContent(selEntry[3])
	}
	return m.styles.viewport.Render(lipgloss.JoinVertical(lipgloss.Left, headerStyle.Render(m.tab.eView.header), m.tab.eView.viewPort.View()))
}
