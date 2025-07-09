package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) searchInit() tea.Cmd {
	m.tab.tiS = textinput.New()
	m.tab.tagS = textinput.New()
	m.tab.daS = textinput.New()
	searchBoxStyle.Width(m.width / 2)
	m.tab.tiS.Placeholder = "search with title"
	m.tab.tagS.Placeholder = "search with tags"
	m.tab.daS.Placeholder = "search with date"

	m.tab.tiS.Width = 20
	m.tab.tagS.Width = 20
	m.tab.daS.Width = 20

	return nil
}

func (m *model) searchUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {

	var p *textinput.Model // pointer to the selected textinput
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.Type {
		case tea.KeyDown:
			if m.tab.cursor != 3 {
				m.tab.cursor++
			}
		case tea.KeyUp:
			if m.tab.cursor != 0 {
				m.tab.cursor--
			}
		}
	}

	switch m.tab.cursor {
	case 0:
		p = &m.tab.tiS

	case 1:
		p = &m.tab.daS
	case 2:
		p = &m.tab.tagS
	}

	// Focus the selected input and blur the others
	m.tab.tiS.Blur()
	m.tab.daS.Blur()
	m.tab.tagS.Blur()
	p.Focus()

	//final
	updated, cmd := p.Update(msg)
	*p = updated

	m.filter(m.tab.tiS.Value(), m.tab.daS.Value(), m.tab.tagS.Value())
	//m.tab.table.SetRows(m.tab.filteredRows.rows)
	return m, cmd
}

func (m *model) searchView() string {

	return searchBoxStyle.Render(
		"search options: \n",
		"search title:",
		m.tab.tiS.View(),
		"\n search date:",
		m.tab.daS.View(),
		"\n search tags",
		m.tab.tagS.View(),
	)

}

func deepCopyRows(rows []table.Row) []table.Row {
	newRows := make([]table.Row, len(rows))
	for i, row := range rows {
		newRow := make(table.Row, len(row))
		copy(newRow, row)
		newRows[i] = newRow
	}
	return newRows
}

// helper function to filter table entries
func (m *model) filter(title string, date string, tags string) {

	rows := deepCopyRows(m.tab.rows.rows)
	data := make([]hiddenData, len(m.tab.rows.rows))
	copy(m.tab.rows.data, data)
	// Return original if all filters are empty
	if title == "" && date == "" && tags == "" {
		m.tab.filteredRows = m.tab.rows
		m.tab.table.SetRows(m.tab.rows.rows)
		return
	}

	var filtered rowData
	for i, val := range rows {
		keep := true

		if title != "" && !strings.Contains(val[0], title) {
			keep = false
		}
		if date != "" && !strings.Contains(val[1], date) {
			keep = false
		}
		if tags != "" && !strings.Contains(val[2], tags) {
			keep = false
		}

		if keep {
			filtered.rows = append(filtered.rows, val)
			filtered.data = append(filtered.data, data[i])
		}
	}

	m.tab.filteredRows = filtered
	m.tab.table.SetRows(filtered.rows)
}
