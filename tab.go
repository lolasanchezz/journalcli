package main

import (
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
	var width = 25
	columns := []table.Column{
		{Title: "title", Width: width},
		{Title: "date written", Width: width},
		{Title: "tags", Width: width},
		{Title: "hidden", Width: 0},
	}

	// Show loading row first
	m.tab.table = table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{{"loading rows in", "", ""}}),
		table.WithHeight(m.height/4),
	)

	// Set table styles
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

	// Batch loading state + actual data load
	return tea.Batch(
		setLoading,
		func() tea.Msg {
			newData, err := takeOutData(m.pswdUnhashed, m.secretsPath)
			if err != nil {
				m.errMsg = err
				return dataLoadedIn{
					data: jsonEntries{readIn: 1},
					rows: rowData{rows: []table.Row{{"error loading data", "", "", ""}}},
				}
			}

			if len(newData.Entries) == 0 {
				return dataLoadedIn{
					data: jsonEntries{readIn: 1},
					rows: rowData{rows: []table.Row{{"no data yet!", "", "", ""}}},
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
				}
				rows.data[i] = hiddenData{
					obj.Msg,
					i, //an id
					//so you can IMMEDIATELY see an entries message
				}
			}

			return dataLoadedIn{
				data: newData,
				rows: rows,
			}
		},
	)
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
