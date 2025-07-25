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
	inputs      [6]textinput.Model
	inputval    string
	inputNames  [6]string
}
type conf struct {
	JournalHash  string  `json:"JournalHash"`
	TextColor    string  `json:"TextColor"`
	BordCol      string  `json:"BordCol"`
	SecTextColor string  `json:"SecTextColor"`
	Width        float64 `json:"Width"`
	Height       float64 `json:"Height"`
	Fullscreen   bool    `json:"Fullscreen"`
}

func (m model) settingsInit() (model, tea.Cmd) {
	//load in existing settings
	existConf, err := takeOutConfig(m.confPath)
	if err != nil {
		m.errMsg = err
		return m, nil
	}
	m.settings.inputNames = [6]string{"border color", "text color", "secondary text color", "% of terminal width", "% of terminal height", "fullscreen"}
	m.settings.inputs = [6]textinput.Model{}

	m.settings.currentConf = existConf
	m.settings.cursor = 0
	m.settings.inputs[0] = textinput.New()
	m.settings.inputs[0].Placeholder = existConf.BordCol
	m.settings.inputs[0].SetValue(existConf.BordCol)
	m.settings.inputs[1] = textinput.New()
	m.settings.inputs[1].Placeholder = existConf.TextColor
	m.settings.inputs[1].SetValue(existConf.TextColor)
	m.settings.inputs[2] = textinput.New()
	m.settings.inputs[2].Placeholder = existConf.SecTextColor
	m.settings.inputs[2].SetValue(existConf.SecTextColor)
	m.settings.inputs[3] = textinput.New()
	m.settings.inputs[3].Placeholder = strconv.FormatFloat(existConf.Width, byte('f'), 4, 64)
	m.settings.inputs[3].SetValue(strconv.FormatFloat(existConf.Width, byte('f'), 4, 64))
	m.settings.inputs[4] = textinput.New()
	m.settings.inputs[4].Placeholder = strconv.FormatFloat(existConf.Height, byte('f'), 4, 64)
	m.settings.inputs[4].SetValue(strconv.FormatFloat(existConf.Height, byte('f'), 4, 64))
	m.settings.inputs[5] = textinput.New()
	m.settings.inputs[5].Placeholder = strconv.FormatBool(existConf.Fullscreen)
	m.settings.inputs[5].SetValue(strconv.FormatBool(existConf.Fullscreen))

	m.settings.inputs[0].Focus()

	//just to make my life easier

	for i := range len(m.settings.inputs) {
		m.settings.inputs[i].Width = lipgloss.Width(m.settings.inputs[i].Value())
	}

	return m, nil
}

func (m model) settingsUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
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
		case tea.KeyCtrlS:

			width, err := strconv.ParseFloat(m.settings.inputs[3].Value(), 64)
			if (err != nil) || (width > 0.95) {
				m.settings.inputval = "width val must be between 0-0.95"
				return m, nil
			}
			height, err := strconv.ParseFloat(m.settings.inputs[4].Value(), 64)
			if err != nil || (height > 0.95) {
				m.settings.inputval = "height val must be between 0-0.95"
				return m, nil
			}
			for i := range m.settings.inputs {
				if len(m.settings.inputs[i].Value()) == 0 {
					m.settings.inputs[i].SetValue(m.settings.inputs[i].Placeholder)
				}
				if i < 3 {
					if rune(m.settings.inputs[i].Value()[0]) != '#' || len(m.settings.inputs[i].Value()) != 7 {
						m.settings.inputval = "invalid rgb color"
						return m, nil
					}
				}
			}

			fs, err := strconv.ParseBool(m.settings.inputs[5].Value())
			if err != nil {
				m.settings.inputval = "fullscreen must be yes or no"
			}

			if fs {
				cmd = tea.EnterAltScreen
				m.aHeight = m.width - 2
				m.aWidth = m.width - 2

			} else {
				cmd = tea.ExitAltScreen
			}
			newConf := conf{
				JournalHash:  m.pswdHash,
				TextColor:    m.settings.inputs[1].Value(),
				BordCol:      m.settings.inputs[0].Value(),
				SecTextColor: m.settings.inputs[2].Value(),
				Width:        width,
				Height:       height,
				Fullscreen:   fs,
			}
			putInConfig(m.confPath, newConf)
			m.config = newConf

			m.settings.inputval = ""

			m.action = 1
			m.settings.cursor = 0

			m.styles.header = m.styles.header.Foreground(lipgloss.Color(m.settings.inputs[2].Value()))
			m.styles.filter = m.styles.filter.BorderForeground(lipgloss.Color(m.settings.inputs[0].Value())).Foreground(lipgloss.Color(m.settings.inputs[1].Value()))
			m.styles.root = m.styles.root.Foreground(lipgloss.Color(m.settings.inputs[1].Value())).
				BorderForeground(lipgloss.Color(m.settings.inputs[0].Value())).
				Width(int(float64(m.width) * newConf.Width)).
				Height(int(float64(m.height) * newConf.Height))

			return m, cmd
		}

	}

	m.settings.inputs[m.settings.cursor].Focus()
	m.settings.inputs[m.settings.cursor], cmd = m.settings.inputs[m.settings.cursor].Update(msg)
	//m.settings.Height.Focus()
	//	m.settings.Height, cmd = m.settings.Height.Update(msg)
	return m, cmd
}

func (m *model) settingsView() string {

	var inputTxt = make([]string, 8)
	for i := range len(inputTxt) - 2 {
		inputTxt[i+2] = lipgloss.JoinHorizontal(lipgloss.Center, m.settings.inputNames[i], "  ", m.settings.inputs[i].View())
	}
	inputTxt[0] = "settings"
	inputTxt[1] = m.settings.inputval
	str := lipgloss.JoinVertical(lipgloss.Center, inputTxt...)
	return str

}
