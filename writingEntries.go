package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type writing struct {
	titleInput textinput.Model
	tagInput   textinput.Model
	existEntry entry
	entryId    int //zero if nothing, index + 1 if we are editing an existing entry
	body       textarea.Model
	typingIn   int
	tagStr     string
}

//for switching between text inputs

func (m *model) writeInit() tea.Cmd {
	//just setting up inputs no biggie
	m.entryView.titleInput = textinput.New()
	m.entryView.tagInput = textinput.New()
	m.entryView.body = textarea.New()

	//formatting
	m.entryView.titleInput.CharLimit, m.entryView.tagInput.CharLimit = 156, 156
	m.entryView.body.SetWidth(int(float64(m.width) * 0.7))

	m.entryView.tagInput.Width, m.entryView.titleInput.Width = 50, lipgloss.Width(time.Now().Format(timeFormat))
	//placeholders!
	m.entryView.titleInput.Placeholder = time.Now().Format(timeFormat)
	m.entryView.tagInput.Placeholder = "tags..."
	//now also need to fetch tags from file. will use cmd for this

	if m.data.readIn == 0 {
		m.loading = true
		return tea.Sequence(m.entryView.tagInput.Focus(), setLoading, tea.Cmd(func() tea.Msg {
			// attempt to fetch data
			tmp, err := takeOutData(m.pswdUnhashed, m.secretsPath)
			if err != nil {
				m.errMsg = err
				return nil
			}

			if tmp.readIn == 1 { //means theres something in the file
				//sort through tags
				var tags string

				if len(tmp.Tags) > 0 {
					tags = "["
					l := len(tmp.Tags)
					for i := range tmp.Tags {
						l--
						if l == 0 {
							tags += i + "]"
						} else {
							tags += i + ", "
						}
					}
				} else {
					tags = "none yet!"
				}
				m.entryView.tagStr = tags
				return dataLoadedIn{data: tmp, msg: tags}
			} else {
				//nothing in the file
				return dataLoadedIn{data: jsonEntries{readIn: 1, Tags: make(map[string]int)}}
			}

		}))
	} else { //data is read in , make tags
		var tags string

		if len(m.data.Tags) > 0 {
			tags = "["
			l := len(m.data.Tags)
			for i := range m.data.Tags {
				l--
				if l == 0 {
					tags += i + "]"
				} else {
					tags += i + ", "
				}
			}
		} else {
			tags = "none yet!"

		}
		m.entryView.tagStr = tags
	}
	return nil

}

func uniqueArrMap(bigmap map[string]int, slices ...[]string) map[string]int {
	var all []string
	for _, slice := range slices {
		all = append(all, slice...)
	}

	for _, v := range all {
		v = strings.TrimSpace(v)
		v = strings.ToLower(v)
		bigmap[v] = bigmap[v] + 1
	}
	return bigmap
}

type dataLoadedIn struct {
	data jsonEntries
	rows rowData
	msg  string
	msgi int
}

func (m model) writingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd = nil
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case dataLoadedIn:
		m.entryView.tagStr = msg.msg
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlS:

			//load in data. decrypt it. add most recent entry. encrypt it. put it back
			m.action = 1
			m.saving = true
			//decrypting part!
			//since this returns nothing if the file is empty or doesn't exist, we don't have to worry about other error handling
			//running an io here
			return m,
				tea.Sequence(m.entryView.titleInput.Focus(), setLoading, tea.Cmd(
					func() tea.Msg {
						var msg dataLoadedIn
						if m.data.readIn == 0 {
							tmp, err := takeOutData(m.pswdUnhashed, m.secretsPath)

							if err != nil {
								m.errMsg = err
							}
							msg.data = tmp
						}
						msg.data = m.data
						debug(msg)
						//new part - load in json tags, seperate new tags by commas, see if there's any new ones not in json
						//add those new ones to json, then take tags from entry and add them to the json entry!

						//now, have to check whether we are writing a new value or were passed one from readingEntries.
						var pastEntries []entry
						var unique map[string]int
						titleStr := m.entryView.titleInput.Value()

						var newTags []string
						if m.entryView.tagInput.Value() == "" {
							newTags = []string{""}
						} else {
							newTags = strings.Split(m.entryView.tagInput.Value(), ",")
						}
						if titleStr == "" { //if no title was entered
							titleStr = m.entryView.titleInput.Placeholder
						}
						if newTags[0] != "" {
							//update unique map
							if m.entryView.tagInput.Value() != "" {

								newTags := strings.Split(m.entryView.tagInput.Value(), ",")
								unique = uniqueArrMap(msg.data.Tags, newTags)
								msg.data.Tags = unique
							}
						}
						if m.entryView.entryId != 0 {
							oldEntry := m.entryView.existEntry
							if titleStr != "" {
								oldEntry.Title = m.entryView.titleInput.Value()
							} else {
								oldEntry.Title = "edit on" + time.Now().Format(timeFormat)
							}

							oldEntry.Msg = m.entryView.body.Value()
							oldEntry.Tags = newTags
							pastEntries = msg.data.Entries
							pastEntries[m.entryView.entryId-1] = oldEntry
						} else {
							pastEntries = append(msg.data.Entries, entry{Title: titleStr, Msg: m.entryView.body.Value(), Date: time.Now(), Tags: newTags})
						}
						//add past entries for viewing
						msg.data.Entries = pastEntries
						debug(msg)
						putInFile(msg.data, m.pswdUnhashed, m.secretsPath)

						return msg //usually would have to do something with this, but because you can write an entry and
						// just exit and have it save while you look through the list, its no biggie and nothing needs to signify to the user
						//that its saving
					},
				))

		case tea.KeyEsc:
			m.action = 1

		case tea.KeyUp:

			if (m.entryView.typingIn != 0) && m.entryView.body.Line() == 0 {
				m.entryView.typingIn--

			}

		case tea.KeyDown:
			if m.entryView.typingIn != 2 {
				m.entryView.typingIn++
			}
		}

		//the responding text input correlates to whatever the "typing in" int is

	}

	if m.entryView.typingIn == 0 { //on title

		//have to this every time .. //TODO there is definitely a better way
		m.entryView.tagInput.Cursor.Blur()
		m.entryView.body.Cursor.Blur()

		m.entryView.titleInput.Focus()
		m.entryView.titleInput, cmd = m.entryView.titleInput.Update(msg)
		return m, cmd
	}

	if m.entryView.typingIn == 1 { //on tags

		m.entryView.titleInput.Cursor.Blur()
		m.entryView.body.Cursor.Blur()

		m.entryView.tagInput.Focus()
		m.entryView.tagInput, cmd = m.entryView.tagInput.Update(msg)
		return m, cmd
	}

	if m.entryView.typingIn == 2 { //on body writing

		m.entryView.titleInput.Cursor.Blur()
		m.entryView.tagInput.Cursor.Blur()

		m.entryView.body.Focus()
		m.entryView.body, cmd = m.entryView.body.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	}

	return m, cmd

}

func (m *model) writingView() string {

	m.entryView.body.SetWidth(int(float64(m.width) * 0.7))

	var tags string
	if m.loading {
		tags = " loading...."
	} else if m.entryView.tagStr == "" {
		tags = "none yet!"
	} else {
		tags = m.entryView.tagStr
	}
	return lipgloss.JoinVertical(lipgloss.Center,
		("title:" +
			m.entryView.titleInput.View()),
		("tags (seperate by comma)" +
			m.entryView.tagInput.View()),

		("past tags:" +
			tags),
		"write entry below!",
		m.entryView.body.View(),
		"esc to go back, ctrl + c to quit")
}
