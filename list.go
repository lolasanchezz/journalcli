package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type picking struct {
	choices []string
	cursor  int
	spinner spinner.Model
}

func (m model) listInit() (model, tea.Cmd) {
	m.list.choices = []string{"write entries", "read entries", "change password", "look at analytics", "settings", "logout"}

	m.list.cursor = 0
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m.list.spinner = s

	if m.loading {
		return m, m.list.spinner.Tick
	}
	return m, nil
}

func (m model) listUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case loading:
		m.loading = true
		return m, m.list.spinner.Tick
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.action == 1 {
				if m.list.cursor > 0 {
					m.list.cursor--
				}
			}
		case tea.KeyDown:
			if m.action == 1 {
				if m.list.cursor < len(m.list.choices) {
					m.list.cursor++
				}
			}
		case tea.KeyEnter:
			m.action = m.list.cursor + 2
		}
	default:

	}
	if m.loading {
		var spinnerCmd tea.Cmd
		m.list.spinner, spinnerCmd = m.list.spinner.Update(msg)
		cmd = tea.Batch(cmd, spinnerCmd)
	}
	//setting up each model for when the action is clicked
	if m.action == 2 { //writing a new entry!
		return m, m.writeInit()
	}
	if m.action == 3 {
		return m, m.readInit()
	}
	if m.action == 4 {
		return m, m.psrsInit()
	}
	if m.action == 7 {
		return m, tea.Quit
	}

	return m, cmd

}

func (m *model) listView() string {
	var s string
	var sArr []string
	sArr = append(sArr, "what would you like to do?")
	for i, choice := range m.list.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.list.cursor == i {
			cursor = "> " // cursor!
		}

		// Render the row
		sArr = append(sArr, (cursor + choice))
	}
	if m.saving {
		sArr = append(sArr, ("saving entry " + m.list.spinner.View()))
	}

	len := len(sArr)
	reverseArr := make([]string, len)
	var index int
	for i, val := range sArr {
		index = len - (i + 1)
		reverseArr[index] = val
	}

	s = lipgloss.JoinVertical(lipgloss.Center, sArr...)
	return s
}
