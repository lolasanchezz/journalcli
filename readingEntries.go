package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewDat struct {
	table    table.Model
	view     int //for choosing between either scrolling through table or going through modules
	maxViews int
	tiS      textinput.Model
	daS      textinput.Model
	tagS     textinput.Model
	cursor   int         //for selecting between searching
	rows     []table.Row //changes as search fields change
}

func (m *model) readInit() tea.Cmd {
	m.tabInit()
	m.searchInit()
	return nil
}

func (m *model) tabInit() tea.Cmd {
	var rows []table.Row
	var width = 25
	columns := []table.Column{{Title: "title", Width: width}, {Title: "date written", Width: width}, {Title: "tags", Width: width}}
	//if data hasn't been decrypted yet (if no entry has been written)

	if newData, err := takeOutData(m.pswdUnhashed, m.secretsPath); len(newData.Entries) == 0 { //if no data available
		if err != nil {
			m.errMsg = err
		}
		rows = []table.Row{{"no entries yet!"}}

	} else {

		rows = make([]table.Row, len(newData.Entries))

		for index, obj := range newData.Entries {
			tagStr := strings.Join(obj.Tags, ", ")
			rows[index] = table.Row{obj.Title, obj.Date.Format(timeFormat), tagStr}
		}
	}
	m.tab.rows = rows
	m.tab.table = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	//rot styling copied from docs
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.tab.table.SetStyles(s)

	return nil

}

func (m *model) readUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width

		// Update column widths based on new width
		// Reserve room for spacing, search box, etc.
		usableWidth := m.width / 2
		numCols := len(m.tab.table.Columns())
		colWidth := usableWidth / numCols

		cols := m.tab.table.Columns()
		for i := range cols {
			cols[i].Width = colWidth
		}
		m.tab.table.SetColumns(cols)

		return m, nil

	case tea.KeyMsg:

		switch msg.Type {
		case tea.KeyEsc:
			m.action = 1
		case tea.KeyRight:
			if m.tab.view == 0 {
				m.tab.view++
			}

		case tea.KeyLeft:
			if m.tab.view == 1 {
				m.tab.view--
			}

		}

	} //otherwise, just pass onto helping functions
	if m.tab.view == 0 {
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
	return m, nil
}

func (m *model) readView() string {
	return lipgloss.JoinHorizontal(lipgloss.Top, m.tab.table.View(), m.searchView())
}

func (m model) tabUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// default for now
	m.tab.table, cmd = m.tab.table.Update(msg)
	return m, cmd
}

var searchBoxStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")).
	AlignHorizontal(lipgloss.Left).
	Border(lipgloss.ThickBorder(), true, true)

// making the searching aspects on the side
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
	rows1 := deepCopyRows(m.tab.rows)

	// Return original if all filters are empty
	if title == "" && date == "" && tags == "" {
		m.tab.table.SetRows(m.tab.rows)
		return
	}

	var filtered []table.Row
	for _, val := range rows1 {
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
			filtered = append(filtered, val)
		}
	}

	m.tab.table.SetRows(filtered)
}
