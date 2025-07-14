package main

import (
	"github.com/charmbracelet/lipgloss"
)

type styles struct {
	light      string  //light color
	dark       string  //dark color
	first      string  //first/text color
	second     string  //secondary text color
	widthPerc  float64 //percentage of terminal width
	heightPerc float64 //percentage of terminal height
}

var (

	//default colors
	light  = "#699ACC"
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
			Padding(1)

	headerStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Center).
			Italic(true).
			Bold(true)

	searchBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(light)).
			AlignHorizontal(lipgloss.Left).
			Border(lipgloss.ThickBorder(), true, true).
			Padding(1)
)
