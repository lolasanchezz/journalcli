package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (

	//colors
	blue   = "#699ACC"
	green  = "#61AA07"
	pwhite = "#F4DBCC"
	dgreen = "#394032"

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
			BorderForeground(lipgloss.Color(blue)).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color(pwhite)).
			Padding(1).
			Width(20).
			AlignVertical(lipgloss.Center)

	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lineBorder).
			Foreground(lipgloss.Color(green)).
			AlignHorizontal(lipgloss.Center)

	headerStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Center).
			Italic(true).
			Bold(true)

	searchBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(pwhite)).
			AlignHorizontal(lipgloss.Left).
			Border(lipgloss.ThickBorder(), true, true).
			Padding(1)
)
