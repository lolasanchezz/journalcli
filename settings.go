package main

import (
	"github.com/charmbracelet/bubbles/textinput"
)

//things to modify ->
// border color
// text color
// secondary color
// default height
// default width
// fullscreen

type settingInp struct {
	bordCol      textinput.Model
	textColor    textinput.Model
	secTextColor textinput.Model
}
