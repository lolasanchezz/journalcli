package main

import (
	"strings"

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
		finalStr = " " + row[3]
	} else {
		finalStr = fitLinesLipgloss(row[3], m.aWidth/2-2)
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
	/*
		selEntry := m.tab.table.SelectedRow()

		if selEntry != nil {

			if len(selEntry) == 0 {
				m.tab.eView.viewPort.SetContent("")
			} else {
				m.setRow(selEntry)
			}

		}
	*/
	return m.styles.viewport.Render(lipgloss.JoinVertical(lipgloss.Center, headerStyle.Render(m.tab.eView.header), "\n",
		m.tab.eView.viewPort.View()))
}

func (m *model) updateViewportCont() {
	if !m.tab.loading {
		selEntry := m.tab.table.SelectedRow()

		if selEntry != nil {

			if len(selEntry) == 0 {
				m.tab.eView.viewPort.SetContent("")
			} else {
				m.setRow(selEntry)
			}

		}
	}
}

func (m *model) resizeViewport() {
	width := m.aWidth/2 - 4
	height := m.aHeight - 6

	m.styles.viewport = m.styles.viewport.Width(width).Height(height)
	m.tab.eView.viewPort.Width = width
	m.tab.eView.viewPort.Height = height

}

func fitLines(body string, w int) string {
	padding := " "
	var finalStr string
	firstspace := ""
	secondspace := ""
	mod := 0
	pmod := 0

	for i := range int((len(body)) / (w)) {
		firstspace = ""
		secondspace = ""
		pmod = mod
		mod = 2
		if body[w*i] != ' ' {
			firstspace = padding
			mod--
		}

		if w*(i+1) >= len(body) {
			finalStr += firstspace + body[w*i:]
			break // we're done
		}
		if body[w*(i+1)] != ' ' {
			secondspace = padding
			mod--
		}
		if strings.Contains(body[w*i:w*i+1], "\n") {
			finalStr += body[w*i:]
			//line is already broken no need to format it
		} else {

			finalStr += firstspace + body[w*i+pmod:w*(i+1)+mod] + secondspace + " \n"
		}

	}
	return finalStr
}

func fitLinesLipgloss(body string, w int) string {
	return lipgloss.NewStyle().Width(w).Padding(0, 1).AlignHorizontal(lipgloss.Left).Render(body)

}
