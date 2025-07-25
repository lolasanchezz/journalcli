package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var defaultStyles = conf{
	TextColor:    "#F4DBCC",
	BordCol:      "#F4DBCC",
	SecTextColor: "#F4DBCC",
	Width:        0.9,
	Height:       0.5,
}

type styles struct {
	root     lipgloss.Style
	viewport lipgloss.Style
	header   lipgloss.Style
	filter   lipgloss.Style
}

func (m *model) setStyles(new conf) {
	m.styles.root.BorderForeground(lipgloss.Color(new.BordCol))
	m.styles.root.Foreground(lipgloss.Color(new.TextColor))
	m.styles.viewport.Foreground(lipgloss.Color(new.TextColor))
	m.styles.filter.Foreground(lipgloss.Color(new.TextColor))

}

var (

	//default colors
	light  = "#F4DBCC"
	dark   = "#61AA07"
	first  = "#F4DBCC"
	second = "#394032"
	border = light

	//other presets
	lineBorder = lipgloss.Border{
		Top:    "~",
		Left:   "|",
		Right:  "|",
		Bottom: "^",
	}

	inlinePadding = lipgloss.NewStyle().Padding(1)

	//styles
	rootStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(border)).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color(first)).
			Padding(1).
			Width(20).
			AlignVertical(lipgloss.Center)

	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lineBorder).
			Foreground(lipgloss.Color(light)).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Padding()

	headerStyle = lipgloss.NewStyle().
			Italic(true).
			Bold(true)

	searchBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(light)).
			AlignHorizontal(lipgloss.Left).
			Border(lipgloss.ThickBorder(), true, true).
			Padding(1)
)

//just some helper funcs

func (m *model) checkWidth(ws ...int) bool {
	var totalW int
	for _, val := range ws {
		totalW += val
	}
	return totalW > m.styles.root.GetWidth()-3

}

func (m *model) addHelp(str string) string {
	maxHeight := float64(m.height) * m.config.Height
	maxWidth := float64(m.width) * m.config.Width
	m.help.Width = int(maxWidth)
	lineNum := maxHeight - ((maxHeight / 2) + (float64(lipgloss.Height(str))) - 2)
	if lineNum < 0 {
		return str
	}
	return lipgloss.JoinVertical(lipgloss.Center, str, strings.Repeat("\n", int(lineNum)), m.help.View(keys))
}
