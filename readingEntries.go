package main

import (
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewDat struct {
	table        table.Model
	view         int //for choosing between either scrolling through table or going through modules
	tiS          textinput.Model
	daS          textinput.Model
	tagS         textinput.Model
	cursor       int     //for selecting between searching
	rows         rowData //changes as search fields change
	filteredRows rowData
	viewsEnabled []bool
	maxViews     int
	eView        viewLog
}

func (m *model) readInit() tea.Cmd {
	m.tab.viewsEnabled = []bool{true, true, false}
	m.tab.maxViews = 2
	m.searchInit()
	return m.tabInit()

}

func (m *model) readView() string {

	str := lipgloss.JoinHorizontal(lipgloss.Top, m.tab.table.View(), m.searchView())
	if m.tab.maxViews == 3 {
		return lipgloss.JoinVertical(lipgloss.Bottom, m.viewportView(), str)
	}
	return str
}

func (m *model) readUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case dataLoadedIn:
		m.tab.rows = msg.rows
		m.tab.table.SetRows(msg.rows.rows)
		m.tab.table.Focus()
	case tea.WindowSizeMsg:
		// Update column widths based on new width
		// Reserve room for spacing, search box, etc.
		usableWidth := m.width / 2
		hiddenCols := 2
		numCols := len(m.tab.table.Columns()) - hiddenCols //have to exclude two hidden columns
		colWidth := usableWidth / numCols

		cols := m.tab.table.Columns()
		for i := range cols {
			if i >= numCols {
				break

			}
			cols[i].Width = colWidth
		}
		m.tab.table.SetColumns(cols)

		return m, nil

	case tea.KeyMsg:

		switch msg.String() {
		case "esc":
			m.action = 1
		case "right":
			if m.tab.view != (m.tab.maxViews - 1) {
				m.tab.view++
			}

		case "left":
			if m.tab.view != 0 {
				m.tab.view--
			}

		case "v":
			if !(m.tab.view == 2) {
				if m.tab.maxViews == 3 {
					m.tab.maxViews = 2
					m.tab.viewsEnabled[2] = false

				} else {
					m.tab.maxViews = 3
					m.tab.viewsEnabled[2] = true
					return m.viewportInit(), nil
				}

			}

		case "enter":
			//switch over to writing
			m.writeInit()
			var e entry
			row := m.tab.table.SelectedRow()
			e.Date, _ = time.Parse(timeFormat, row[1])
			i, _ := strconv.Atoi(row[4])
			m.entryView.entryId = i + 1
			m.entryView.existEntry = e

			m.entryView.tagInput.SetValue(row[2])
			m.entryView.titleInput.SetValue(row[0])
			m.entryView.body.SetValue(row[3])

			m.action = 2

		}

	} //otherwise, just pass onto helping functions

	if m.tab.viewsEnabled[2] {
		m.viewportUpdate(msg) //need to update this no matter what
	}

	if m.tab.view == 0 { //table open - always open!
		m.tab.tiS.Blur()
		m.tab.daS.Blur()
		m.tab.tagS.Blur()
		m.tab.table.Focus()
		return m.tabUpdate(msg)
	}
	if m.tab.view == 1 {
		m.tab.table.Blur()
		return m.searchUpdate(msg)
	}
	if m.tab.view == 2 {
		return m.viewportUpdate(msg)
	}

	return m, nil
}
