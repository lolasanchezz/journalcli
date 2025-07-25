package main

import (
	"math"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
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

	//splitting the content by line breaks
	var finalStr string
	if len(row[3]) < m.aWidth/2 {
		finalStr = row[3]
	} else {
		realW := m.aWidth / 2
		for i := range int(math.Ceil(float64(len(row[3])) / (float64(realW)))) {
			if realW*(i+1) > len(row[3]) {
				finalStr += row[3][realW*i:]
				break
			}
			finalStr += row[3][realW*i : realW*(i+1)]
			finalStr += "\n"
		}

	}
	m.tab.eView.viewPort.SetContent(finalStr)
}

func (m model) viewportInit() model {

	m.tab.eView.viewPort = viewport.New(m.aWidth/2, m.aHeight-5)
	//we can assume that data is already loaded in, as this functino shouldn't even be triggered if data isn't loaded into table
	selEntry := m.tab.table.SelectedRow()

	if len(selEntry) == 0 {
		m.tab.eView.viewPort.SetContent("")
		return m
	}
	m.setRow(selEntry)
	return m
}

func (m *model) viewportView() string {

	selEntry := m.tab.table.SelectedRow()

	if selEntry != nil {

		if len(selEntry) == 0 {
			m.tab.eView.viewPort.SetContent("")
		} else {
			m.setRow(selEntry)
		}

	}
	return m.styles.viewport.Render(lipgloss.JoinVertical(lipgloss.Left, headerStyle.Render(m.tab.eView.header), m.tab.eView.viewPort.View()))
}

func (m *model) resizeViewport() {
	width := m.aWidth / 2
	height := m.aHeight - 6

	m.styles.viewport = m.styles.viewport.Width(width).Height(height)
	m.tab.eView.viewPort.Width = width
	m.tab.eView.viewPort.Height = height

}
