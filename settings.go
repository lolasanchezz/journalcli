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
	BordCol      textinput.Model
	TextColor    textinput.Model
	SecTextColor textinput.Model
	Width        textinput.Model
	Height       textinput.Model
	cursor       int
	currentConf  conf
	inputs       []*textinput.Model
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
	//just to make my life easier

	m.settings.currentConf = existConf
	m.settings.cursor = 0
	m.settings.BordCol = textinput.New()
	m.settings.BordCol.Placeholder = "current border color: " + existConf.BordCol
	m.settings.TextColor = textinput.New()
	m.settings.TextColor.Placeholder = "current text color: " + existConf.TextColor
	m.settings.SecTextColor = textinput.New()
	m.settings.SecTextColor.Placeholder = "current secondary text color: " + existConf.SecTextColor
	m.settings.Width = textinput.New()
	m.settings.Width.Placeholder = "current width % of terminal: " + strconv.FormatFloat(existConf.Width, byte(0), 4, 64)
	m.settings.Height = textinput.New()
	m.settings.Height.Placeholder = "current height % of terminal: " + strconv.FormatFloat(existConf.Height, byte(0), 4, 64)

	m.settings.BordCol.Focus()
	m.settings.cursor = 0
	m.settings.inputs = []*textinput.Model{
		&m.settings.TextColor,
		&m.settings.BordCol,
		&m.settings.SecTextColor,
		&m.settings.Width,
		&m.settings.Height,
	}
	for _, val := range m.settings.inputs {
		val.Width = lipgloss.Width(val.Placeholder)
	}

	return m, nil
}

func (m *model) settingsUpdate(msg tea.Msg) (model, tea.Cmd) {

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
			float, _ := strconv.ParseFloat(m.settings.Width.Value(), 64)
			height, err := strconv.ParseFloat(m.settings.Height.Value(), 64)
			if err != nil {
				m.errMsg = err
				return *m, nil
			}
			putInConfig(m.confPath, conf{
				JournalHash:  m.pswdHash,
				TextColor:    m.settings.TextColor.Value(),
				BordCol:      m.settings.BordCol.Value(),
				SecTextColor: m.settings.SecTextColor.Value(),
				Width:        float,
				Height:       height,
			})
		}

	}
	var cmd tea.Cmd
	*m.settings.inputs[m.settings.cursor], cmd = m.settings.inputs[m.settings.cursor].Update(msg)

	return *m, cmd
}

func (m *model) settingsView() string {
	var str string
	for _, val := range m.settings.inputs {
		str = lipgloss.JoinVertical(lipgloss.Center, str, val.View())
	}
	return str
}
