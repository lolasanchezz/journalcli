package main

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var myCuteBorder = lipgloss.Border{
	Top:         "._.:*:",
	Bottom:      "._.:*:",
	Left:        "|*",
	Right:       "|*",
	TopLeft:     "*",
	TopRight:    "*",
	BottomLeft:  "*",
	BottomRight: "*",
}

/*
var aggsBoxStyle = lipgloss.NewStyle().

	Padding(2).
	Width(50).
	AlignVertical(lipgloss.Center).
	AlignHorizontal(lipgloss.Center)
*/
var header = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#ffffff"))

var text = lipgloss.NewStyle().
	Foreground(lipgloss.Color("ffffff"))

func (m *model) viewAggs() string {

	if m.data.readIn == 0 {

		// attempt to fetch data
		tmp, err := takeOutData(m.pswdUnhashed, m.secretsPath)
		if err != nil {
			m.errMsg = err
			return ""
		}
		m.data = tmp
		if tmp.readIn == 0 {
			m.data.readIn = 1
			return header.Render("no data yet!")
		}

	}

	allEntries := m.data.Entries
	sum := len(allEntries)
	var averageLength int

	for _, entry := range allEntries {
		averageLength += len(entry.Msg)

	}
	averageLength = averageLength / sum
	var popTag []string
	var bigNum int
	for tag, num := range m.data.Tags {
		if num > bigNum {
			popTag = []string{tag}
			bigNum = num
		} else if num == bigNum {
			popTag = append(popTag, tag)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header.Render("stats!"),
		text.Render("total entries made: "+strconv.Itoa(sum)),
		text.Render("average char length: "+strconv.Itoa(averageLength)),
		text.Render("most used tag: "+strings.Join(popTag, ", ")),
	)

	return content

}
