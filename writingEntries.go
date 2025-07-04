package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type writing struct {
	titleInput textinput.Model
	tagInput   textinput.Model
	tags       []string
	body       textarea.Model
	typingIn   int
}

//for switching between text inputs

func (m *model) writeInit() {
	//just setting up inputs no biggie
	m.entryView.titleInput = textinput.New()
	m.entryView.tagInput = textinput.New()
	m.entryView.body = textarea.New()

	//placeholders!
	m.entryView.titleInput.Placeholder = time.Now().Format(time.RFC822)
	m.entryView.tagInput.Placeholder = "tags..."
	//now also need to fetch tags from file.
	if m.data.readIn == 0 {
		// attempt to fetch data
		tmp, err := takeOutData(m.pswdUnhashed, m.secretsPath)
		if err != nil {
			m.errMsg = err
			return
		}
		if tmp.readIn == 1 { //means theres something in the file
			m.entryView.tags = tmp.Tags
		} else {
			m.entryView.tags = []string{}
		}
	}
	m.entryView.titleInput.Focus()
}

func (m model) writingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd = nil
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlS:

			//load in data. decrypt it. add most recent entry. encrypt it. put it back

			//decrypting part!
			//since this returns nothing if the file is empty or doesn't exist, we don't have to worry about other error handling
			tmp, err := takeOutData(m.pswdUnhashed, m.secretsPath)
			if err != nil {
				m.errMsg = err
			}
			pastEntries := append(tmp.Entries, entry{Msg: m.entry.textarea.Value(), Date: time.Now()})

			//add past entries for viewing
			m.data.Entries = pastEntries
			//now must reencrypt
			err = putInFile(m.data, m.pswdUnhashed, m.secretsPath)
			if err != nil {
				m.errMsg = err
			}
			m.action = 1

		case tea.KeyEsc:
			m.action = 1

		case tea.KeyUp:
			if m.entryView.typingIn != 0 {
				m.entryView.typingIn--
			}

		case tea.KeyDown:
			if m.entryView.typingIn != 2 {
				m.entryView.typingIn++
			}
		}

	default:
		if !m.entry.textarea.Focused() {
			cmd = m.entry.textarea.Focus()
			cmds = append(cmds, cmd)
			m.entry.textarea, cmd = m.entry.textarea.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		} //this is obviously wrong. just need this to compile

	}
	return m, cmd

}

func (m *model) writingView() string {

	//there should be ->
	// a header text input for an optional title
	// a text input place for tags
	// somehow offer a way to see past tags?
	// like a [tag, tag, tag]
	// a body multiline input
	// help options on the bottom
	// something like this
	// title:(header) 7/3/26 (placeholder for text input)
	// tags: (header) none (placeholder)
	// past tags: [tag, tag, tag]
	// line input

	//make tag line rq
	tags := "["
	for i, val := range m.entryView.tags {
		if i == len(m.entryView.tags)-1 {
			tags += val + "]\n"
		} else {
			tags += val + ", "
		}
	}

	return docStyle.Render(
		"title:",
		m.entryView.titleInput.View(),
		"\n",
		"tags (seperate by comma)",
		m.entryView.tagInput.View(),
		"\n",
		"past tags:",
		tags,
		"write entry here!",
		m.entryView.body.View(),
		"\n esc to go back, ctrl + c to quit",
	)

}
