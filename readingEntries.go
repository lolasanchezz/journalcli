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
	loading      bool
}

func (m *model) readInit() tea.Cmd {
	m.tab.viewsEnabled = []bool{true, true, false}
	m.tab.maxViews = 2
	m.searchInit()
	return m.tabInit()

}

// enter alt screen

//styles

func (m *model) readView() string {

	if m.tab.maxViews == 2 {
		m.sizeTable(0.5)

		//return m.tab.table.View()
		return lipgloss.JoinHorizontal(lipgloss.Center,
			m.tab.table.View(),
			m.searchView(),
		)
	}
	if m.tab.maxViews == 3 {
		tabWid := m.sizeTable(0.4) //resizes width of columns
		m.tab.table.SetHeight(m.aHeight/2 - 4)
		m.styles.filter = m.styles.filter.Width(tabWid)

		//debug(m.tab.table.Width())
		//return lipgloss.JoinVertical(lipgloss.Center, m.searchView(), inlinePadding.Render(m.tab.table.View()))
		return lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.JoinVertical(lipgloss.Center, inlinePadding.Render(m.searchView()), inlinePadding.Render(m.tab.table.View())),

			(m.viewportView()))

	}
	return ""
}

func (m *model) readUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:

		if m.tab.viewsEnabled[2] {
			m.resizeViewport()

		}
	case dataLoadedIn:
		m.tab.rows = msg.rows
		m.tab.table.SetRows(msg.rows.rows)
		m.tab.table.Focus()
		if msg.msgi != 0 {
			m.tab.table.SetHeight(msg.msgi)
		}
		m.tab.loading = false

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
			if !(m.tab.view == 1) && !m.tab.loading {
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
			if !m.loading {
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
				viewportStyle = viewportStyle.BorderForeground(lipgloss.Color(m.config.BordCol))
				searchBoxStyle = searchBoxStyle.UnsetBackground()
				m.action = 2
			}
		}

	} //otherwise, just pass onto helping functions
	//for showing which
	var cmd tea.Cmd
	if m.tab.view == 0 {
		//table open - always open!
		m.tab.tiS.Blur()
		m.tab.daS.Blur()
		m.tab.tagS.Blur()

		m.styles.table.Selected = m.styles.table.Selected.Background(lipgloss.Color(m.config.SecTextColor))
		m.styles.viewport = m.styles.viewport.Foreground(lipgloss.Color(m.config.SecTextColor))

		//have to update viewport
		m.tabUpdate(msg)
		m.updateViewportCont()

	}
	if m.tab.view == 1 {
		m.styles.viewport = m.styles.viewport.Foreground(lipgloss.Color(m.config.SecTextColor))
		m.updateViewportCont()
		//blur table
		m.styles.table.Selected = m.styles.table.Selected.Background(lipgloss.Color(changeRgb(m.config.SecTextColor, 20)))
		return m.searchUpdate(msg)
	}
	if m.tab.view == 2 {
		m.tab.tiS.Blur()
		m.tab.daS.Blur()
		m.tab.tagS.Blur()
		m.styles.table.Selected = m.styles.table.Selected.Background(lipgloss.Color(changeRgb(m.config.SecTextColor, 20)))

		m.styles.viewport = m.styles.viewport.Foreground(lipgloss.Color(m.config.TextColor))
		m.tab.eView.viewPort, cmd = m.tab.eView.viewPort.Update(msg)

	}

	return m, cmd
}

func (m *model) sizeTable(w float64) int {
	usableWidth := float64(m.styles.root.GetWidth()) * w
	// Update column widths based on new width
	// Reserve room for spacing, search box, etc.

	hiddenCols := 2
	numCols := len(m.tab.table.Columns()) - hiddenCols //have to exclude two hidden columns
	colWidth := usableWidth / float64(numCols)

	cols := m.tab.table.Columns()
	for i := range cols {
		if i >= numCols {
			break

		}
		cols[i].Width = int(colWidth)
	}
	m.tab.table.SetColumns(cols)
	//just guessing here and saying that height of table is # columns plus 4
	if (4 + len(m.tab.table.Rows())) > m.aHeight-6 {
		m.tab.table.SetHeight(m.aHeight - 5) //-5 for help
	} else {
		m.tab.table.SetHeight(4 + len(m.tab.table.Rows()))
	}

	m.tab.table.SetStyles(m.styles.table)
	return int(colWidth * float64(numCols))
}
