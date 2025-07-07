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
		m.writeInit()
	}
	if m.action == 7 {
		return m, tea.Quit
	}

	return m, cmd

}

func (m *model) listView() string {
	var s string
	for i, choice := range m.list.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.list.cursor == i {
			cursor = "> " // cursor!
		}

		// Render the row
		s += (cursor + choice + "\n")
	}
	if m.saving {
		s = s + ("\n saving entry " + m.list.spinner.View())
	}
	return docStyle.Render("what would you like to do? \n" + s)

}
