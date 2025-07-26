package main

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var defaultStyles = conf{
	TextColor:    "#F4DBCC",
	BordCol:      "#70391",
	SecTextColor: "#d26323ff",
	Width:        0.9,
	Height:       0.5,
	Fullscreen:   true,
}

type styles struct {
	root     lipgloss.Style
	viewport lipgloss.Style
	header   lipgloss.Style
	filter   lipgloss.Style
	table    table.Styles
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
			AlignVertical(lipgloss.Center)

	headerStyle = lipgloss.NewStyle().
			Italic(true).
			Bold(true)

	searchBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(light)).
			AlignHorizontal(lipgloss.Left).
			Border(lipgloss.ThickBorder(), true, true).
			Padding(2)

	tabStyle = table.DefaultStyles()
)

func (m *model) checkWidth(ws ...int) bool {
	var totalW int
	for _, val := range ws {
		totalW += val
	}
	return totalW > m.styles.root.GetWidth()-3

}

func (m *model) addHelp(str string) string {

	m.help.Width = m.aWidth - 2
	lineNum := m.aHeight - ((m.aHeight / 2) + lipgloss.Height(str)) - 2
	if lineNum < 0 {
		return str
	}
	return lipgloss.JoinVertical(lipgloss.Center, str, strings.Repeat("\n", int(lineNum)), m.help.View(keys))
}

func changeRgb(hex string, light int) string {

	r := hex[1:3]
	g := hex[3:5]
	b := hex[5:7]
	newR, _ := strconv.ParseInt(r, 16, 32)
	newG, _ := strconv.ParseInt(g, 16, 32)
	newB, _ := strconv.ParseInt(b, 16, 32)
	return "#" + strconv.Itoa(int(newR)+light) + strconv.Itoa(int(newG)+light) + strconv.Itoa(int(newB)+light)

}
