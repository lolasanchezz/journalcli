package main

import (
	"crypto/sha256"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pswdEnter struct {
	pswdSet     bool
	pswdWrong   bool
	pswdEntered bool
	ti          textinput.Model
}

func (m model) pswdUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			first := sha256.New()
			if !m.pswdInput.pswdSet {
				//hashing what was just entered and putting it in file
				hash, err := hash(m.pswdInput.ti.Value())
				if err != nil {
					m.errMsg = err
					debug(err)
					return m, tea.Quit
				}
				m.pswdInput.pswdEntered = true
				m.pswdHash = hash

				//now putting that into the file

				err = putInConfig(m.confPath, conf{JournalHash: hash})
				if err != nil {
					m.errMsg = err
					debug(err)
					return m, tea.Quit
				}
				m.pswdHash = hash
				m.pswdUnhashed = m.pswdInput.ti.Value()
				m.pswdInput.ti.Reset()
				m.pswdInput.ti.Focus()
				first.Reset()
				m.listInit()
				m.action = 1
			} else {
				hash, err := hash(m.pswdInput.ti.Value())
				if err != nil {
					m.errMsg = err
					debug(err)
				}
				if hash != m.pswdHash {
					m.pswdInput.pswdWrong = true
					m.pswdInput.ti.Reset()
					m.pswdInput.ti.Focus()
				} else {
					//password is correct!
					m.pswdInput.pswdEntered = true
					m.pswdUnhashed = m.pswdInput.ti.Value()
					m.action = 1
					m.pswdInput.ti.Reset()

					//now we have to prepare the list!

					return m.listInit()

				}
				first.Reset()

			}

		}
	}
	var cmd tea.Cmd
	m.pswdInput.ti, cmd = m.pswdInput.ti.Update(msg)
	return m, cmd
}

func (m model) pswdView() string {
	var fin string
	var header string
	if !m.pswdInput.pswdSet {
		header = "welcome! a password wasn't found in this directory, so enter in a new one!"
	} else if m.pswdInput.pswdWrong && (m.pswdInput.ti.Value() == "") {
		header = "Wrong password! Try again"
	} else {
		m.pswdInput.pswdWrong = false
		header = "Enter in password:"
	}
	header = m.styles.header.Render(header)
	fin = lipgloss.JoinVertical(lipgloss.Center, header, m.pswdInput.ti.View())
	return fin

}
