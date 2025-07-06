package main

import (
	"slices"
	"strings"
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

func (m *model) writeInit() tea.Cmd {
	//just setting up inputs no biggie
	m.entryView.titleInput = textinput.New()
	m.entryView.tagInput = textinput.New()
	m.entryView.body = textarea.New()
	//formatting
	m.entryView.titleInput.CharLimit, m.entryView.tagInput.CharLimit = 156, 156
	m.entryView.titleInput.Width, m.entryView.tagInput.Width = 30, 30
	//placeholders!
	m.entryView.titleInput.Placeholder = time.Now().Format(timeFormat)
	m.entryView.tagInput.Placeholder = "tags..."
	//now also need to fetch tags from file.
	if m.data.readIn == 0 {
		// attempt to fetch data
		tmp, err := takeOutData(m.pswdUnhashed, m.secretsPath)
		if err != nil {
			m.errMsg = err
			return nil
		}
		if tmp.readIn == 1 { //means theres something in the file
			m.entryView.tags = tmp.Tags
		} else {
			m.entryView.tags = []string{}
		}
	}
	m.entryView.titleInput.Focus()
	return m.entryView.titleInput.Focus()
}

func (m model) writingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd = nil
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlS:

			//load in data. decrypt it. add most recent entry. encrypt it. put it back

			//decrypting part!
			//since this returns nothing if the file is empty or doesn't exist, we don't have to worry about other error handling

			if m.data.readIn == 0 {
				tmp, err := takeOutData(m.pswdUnhashed, m.secretsPath)
				if err != nil {
					m.errMsg = err
				}
				m.data = tmp
			}

			//new part - load in json tags, seperate new tags by commas, see if there's any new ones not in json
			//add those new ones to json, then take tags from entry and add them to the json entry!
			newTags := strings.Split(m.entryView.tagInput.Value(), ",")
			var unique []string
			if newTags[0] != "" {
				all := slices.Concat(newTags, m.data.Tags)
				//getting the unique tags
				seen := make(map[string]bool)

				for _, v := range all {
					v = strings.TrimSpace(v)
					v = strings.ToLower(v)
					if !seen[v] { //if the string has been seen already in the map
						seen[v] = true
						unique = append(unique, v)
					}
				}
			} else {
				unique = m.data.Tags
			}
			titleStr := m.entryView.titleInput.Value()
			if titleStr == "" { //if no title was entered
				titleStr = m.entryView.titleInput.Placeholder
			}
			pastEntries := append(m.data.Entries, entry{Title: titleStr, Msg: m.entryView.body.Value(), Date: time.Now(), Tags: newTags})

			//add past entries for viewing
			m.data.Entries = pastEntries
			m.data.Tags = unique
			m.entryView.tags = unique
			//now must reencrypt
			err := putInFile(m.data, m.pswdUnhashed, m.secretsPath)
			if err != nil {
				m.errMsg = err
			}
			m.action = 1

		case tea.KeyEsc:
			m.action = 1

		case tea.KeyLeft:

			if m.entryView.typingIn != 0 {
				m.entryView.typingIn--

			}

		case tea.KeyRight:
			if m.entryView.typingIn != 2 {
				m.entryView.typingIn++
			}
		}

		//the responding text input correlates to whatever the "typing in" int is

	}

	if m.entryView.typingIn == 0 { //on title

		//have to this every time .. //TODO there is definitely a better way
		m.entryView.tagInput.Blur()
		m.entryView.body.Blur()

		m.entryView.titleInput.Focus()
		m.entryView.titleInput, cmd = m.entryView.titleInput.Update(msg)
		return m, cmd
	}

	if m.entryView.typingIn == 1 { //on tags

		m.entryView.titleInput.Blur()
		m.entryView.body.Blur()

		m.entryView.tagInput.Focus()
		m.entryView.tagInput, cmd = m.entryView.tagInput.Update(msg)
		return m, cmd
	}

	if m.entryView.typingIn == 2 { //on body writing

		m.entryView.titleInput.Blur()
		m.entryView.tagInput.Blur()

		m.entryView.body.Focus()
		m.entryView.body, cmd = m.entryView.body.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

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
	var tags string
	if len(m.entryView.tags) > 0 {
		tags = "["
		for i, val := range m.entryView.tags {
			if i == len(m.entryView.tags)-1 {
				tags += val + "]"
			} else {
				tags += val + ", "
			}
		}
	} else {
		tags = "none yet!"
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
		"\nwrite entry below!\n",
		m.entryView.body.View(),
		"\n esc to go back, ctrl + c to quit",
	)

}
