package main

import (
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type hiddenData struct {
	msg string
	id  int
}
type rowData struct {
	rows []table.Row
	data []hiddenData
}

func (m *model) tabInit() tea.Cmd {
	var width = rootStyle.GetWidth() / 3
	columns := []table.Column{
		{Title: "title", Width: width},
		{Title: "date written", Width: width},
		{Title: "tags", Width: width},
		{Title: "hidden", Width: 0},
		{Title: "idhidden", Width: 0},
	}
	//2 for header
	// Show loading row first
	m.tab.table = table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{{"loading rows in", "", ""}}),
		table.WithHeight(5),
	)

	// Set table styles
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(m.config.BordCol)).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color(m.config.TextColor)).
		Background(lipgloss.Color(m.config.SecTextColor)).
		Bold(false)
	m.styles.table = s
	m.tab.table.SetStyles(s)
	m.tab.loading = true
	// Batch loading state + actual data load
	return tea.Batch(
		setLoading,
		func() tea.Msg {

			newData, err := takeOutData(m.pswdUnhashed, m.secretsPath)
			if err != nil {
				m.errMsg = err
				m.tab.loading = false
				return dataLoadedIn{
					data: jsonEntries{readIn: 1},
					rows: rowData{rows: []table.Row{{"error loading data", "", "", ""}}},
					msgi: 3,
				}
			}

			if len(newData.Entries) == 0 {

				return dataLoadedIn{
					data: jsonEntries{readIn: 1},
					rows: rowData{rows: []table.Row{{"no data yet!", "", "", ""}}},
					msgi: 3,
				}
			}
			var rows rowData
			rows.rows = make([]table.Row, len(newData.Entries))
			rows.data = make([]hiddenData, len(newData.Entries))
			for i, obj := range newData.Entries {
				rows.rows[i] = table.Row{
					obj.Title,
					obj.Date.Format(timeFormat),
					strings.Join(obj.Tags, ", "),
					obj.Msg,
					strconv.Itoa(i),
				}
				rows.data[i] = hiddenData{
					obj.Msg,
					i, //an id
					//so you can IMMEDIATELY see an entries message
				}
			}

			//height of table
			tabHeight := len(newData.Entries) + 3
			return dataLoadedIn{
				data: newData,
				rows: rows,
				msgi: tabHeight,
			}
		},
	)
}

func (m *model) tabUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// default for now

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if len(m.data.Entries) == 0 || m.tab.table.Cursor() > len(m.data.Entries) {
				return m, nil
			}
			//first have to delete the tags from the unique map
			row := m.tab.table.SelectedRow()
			if len(row) < 2 {
				return m, nil
			}
			tags := strings.Split(row[2], "")
			if len(tags) > 0 {
				for _, val := range tags {
					if _, ok := m.data.Tags[val]; ok {
						m.data.Tags[val] = m.data.Tags[val] - 1
					}
				}
			}

			//then remove row from data
			i, err := strconv.Atoi(row[4])
			if err != nil {
				m.errMsg = err
				return m, nil
			}
			m.data.Entries = slices.Delete(m.data.Entries, i, i+1)
			cmds = append(cmds, putInFileCmd(m.data, m.pswdUnhashed, m.secretsPath))
			//then remove row from table
			rows := m.tab.table.Rows()
			rowI := 999999
			for i, val := range rows {
				if slices.Equal(val, row) {
					rowI = i
					break
				}
			}
			if rowI == 999999 {
				log.Panic("couldn't find rows in table")
			}
			rows = slices.Delete(rows, rowI, rowI+1)
			//then update index of rows

			for rowI < len(rows) {
				i, _ := strconv.Atoi(rows[rowI][4])
				rows[rowI][4] = strconv.Itoa(i - 1)
				rowI++
			}

			m.tab.table.SetRows(rows)
			//debugging

		}
	}

	m.tab.table, cmd = m.tab.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
