package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pswdReset struct {
	ti textinput.Model
}

type pswdResetMsg struct {
	m model
}

func (m model) psrsInit() tea.Cmd {
	return textinput.Blink
}

func (m model) psrsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlS:

			newPswd := m.psRs.ti.Value()
			if data, err := takeOutData(m.pswdUnhashed, m.secretsPath); len(data.Entries) != 0 {
				if err != nil {
					m.errMsg = err
					return m, nil
				}
				//we know there's data, now we have to reset the password
				err = putInFile(data, newPswd, m.secretsPath)
				if err != nil {
					m.errMsg = err
					return m, nil
				}
			}
			//now writing pswd hash into file
			newHash, err := hash(newPswd)
			m.pswdHash = newHash
			m.pswdUnhashed = newPswd
			if err != nil {
				m.errMsg = err
				return m, nil
			}
			conf := conf{JournalHash: newHash}
			err = putInConfig(m.homeDir, conf)
			if err != nil {
				m.errMsg = err
				return m, nil
			}
			m.listInit()
			m.action = 1
			m.psRs.ti.Reset()

			return m, nil

		}
	}

	m.psRs.ti, cmd = m.psRs.ti.Update(msg)
	return m, cmd
}

func (m model) psrsView() string {
	return lipgloss.JoinVertical(lipgloss.Center, "write new password here:,",
		m.psRs.ti.View(),
		"\n esc to go back, ctrl+s to save",
	)
}
