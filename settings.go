package main

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

//things to modify ->
// border color
// text color
// secondary color
// default height
// default width
// fullscreen

type settingInp struct {
	cursor      int
	currentConf conf
	inputs      [5]textinput.Model
}
type conf struct {
	JournalHash  string  `json:"JournalHash"`
	TextColor    string  `json:"TextColor"`
	BordCol      string  `json:"BordCol"`
	SecTextColor string  `json:"SecTextColor"`
	Width        float64 `json:"Width"`
	Height       float64 `json:"Height"`
}

func (m model) settingsInit() (model, tea.Cmd) {
	//load in existing settings
	existConf, err := takeOutConfig(m.confPath)
	if err != nil {
		m.errMsg = err
		return m, nil
	}

	m.settings.inputs = [5]textinput.Model{}

	m.settings.currentConf = existConf
	m.settings.cursor = 0
	m.settings.inputs[0] = textinput.New()
	m.settings.inputs[0].Placeholder = "current border color: " + existConf.BordCol
	m.settings.inputs[1] = textinput.New()
	m.settings.inputs[1].Placeholder = "current text color: " + existConf.TextColor
	m.settings.inputs[2] = textinput.New()
	m.settings.inputs[2].Placeholder = "current secondary text color: " + existConf.SecTextColor
	m.settings.inputs[3] = textinput.New()
	m.settings.inputs[3].Placeholder = "current width % of terminal: " + strconv.FormatFloat(existConf.Width, byte(0), 4, 64)
	m.settings.inputs[4] = textinput.New()
	m.settings.inputs[4].Placeholder = "current height % of terminal: " + strconv.FormatFloat(existConf.Height, byte(0), 4, 64)

	m.settings.inputs[0].Focus()

	//just to make my life easier

	for i := range len(m.settings.inputs) {
		m.settings.inputs[i].Width = lipgloss.Width(m.settings.inputs[i].Placeholder)
	}

	return m, nil
}

func (m model) settingsUpdate(msg tea.Msg) (model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.settings.cursor > 0 {
				m.settings.inputs[m.settings.cursor].Cursor.Blur()
				m.settings.cursor--
				m.settings.inputs[m.settings.cursor].Focus()

			}
		case tea.KeyDown:
			if m.settings.cursor < len(m.settings.inputs)-1 {
				m.settings.inputs[m.settings.cursor].Cursor.Blur()
				m.settings.cursor++
				m.settings.inputs[m.settings.cursor].Cursor.Focus()

			}
		case tea.KeyCtrlC:
			float, _ := strconv.ParseFloat(m.settings.inputs[4].Value(), 64)
			height, err := strconv.ParseFloat(m.settings.inputs[3].Value(), 64)
			if err != nil {
				m.errMsg = err
				return m, nil
			}
			putInConfig(m.confPath, conf{
				JournalHash:  m.pswdHash,
				TextColor:    m.settings.inputs[0].Value(),
				BordCol:      m.settings.inputs[1].Value(),
				SecTextColor: m.settings.inputs[2].Value(),
				Width:        float,
				Height:       height,
			})
		}

	}
	var cmd tea.Cmd

	m.settings.inputs[m.settings.cursor].Focus()
	m.settings.inputs[m.settings.cursor], cmd = m.settings.inputs[m.settings.cursor].Update(msg)
	//m.settings.Height.Focus()
	//	m.settings.Height, cmd = m.settings.Height.Update(msg)
	return m, cmd
}

func (m *model) settingsView() string {
	str := "settings"
	for _, val := range m.settings.inputs {
		str = lipgloss.JoinVertical(lipgloss.Center, str, val.View())
	}
	return str

}
