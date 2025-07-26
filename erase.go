package main

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//like a lil kill switch

type erase struct {
	ti       textinput.Model
	msg      string
	quitting bool
}

func (m *model) eraseView() string {
	if len(m.erase.msg) > 0 {
		return m.erase.msg
	}
	return lipgloss.JoinVertical(lipgloss.Center,
		"erasing means getting rid of all your data and entries. permanently.",
		m.styles.header.Render("are you sure you want to delete everything?"),
		"type \"i affirm\" and hit enter to delete everything",
		m.erase.ti.View())
}

func (m *model) eraseUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.erase.ti.Value() == "i affirm" {
				//delete everything
				os.Remove(m.secretsPath)
				os.Remove(m.confPath)
				m.erase.msg = "ok! see you later!"
				m.erase.quitting = true
				return m, nil
			}
		}
	}
	if m.erase.quitting {
		return m, tea.Quit
	}
	m.erase.ti, _ = m.erase.ti.Update(msg)
	return m, nil
}

func (m *model) eraseInit() tea.Cmd {
	m.erase.ti = textinput.New()
	m.erase.ti.Width = 14
	m.erase.ti.Placeholder = "are you sure?"
	m.erase.ti.Focus()
	return nil
}
