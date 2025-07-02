package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().
	Margin(1, 2).
	PaddingTop(2).
	PaddingLeft(4).
	Align(lipgloss.Center)

// will take in a password and store the hash in a config file - if no such variable exists, offer to make a new one. the password will
// be the key to decrypt the file with the journal entries. a password is only required to read the past entries, not current
type conf struct {
	JournalHash string `json:"JournalHash"`
}

// setting up the list part
type picking struct {
	choices []string
	cursor  int
}

// typing entries part
type entryWriting struct {
	textarea textarea.Model
}

type jsonEntries struct { //json struct for single entry
	Msg  string    `json:"Msg"`
	Date time.Time `json:"Date"`
}

type viewDat struct {
	table table.Model
}

// TODO: put all the pwsd options in their own struct
type model struct {
	//entering password part

	pswdSet      bool            // does password exist in .env?
	pswdEntered  bool            // has user entered in correct password?
	pswdHash     string          // password in .env
	pswdUnhashed string          // correct password entered in by user (real password)
	pswdWrong    bool            // just showing whether entered in password is incorrect (temporary flash "wrong! in header")
	errMsg       error           //for passing along errors to stdout
	textInput    textinput.Model //text input for password
	//general stuff
	homeDir string //this just gets used so much might as well
	config  conf
	debug   string
	action  int //what r u doing rn?
	//initial list used to select action
	list  picking
	entry entryWriting
	//storing data
	data        []jsonEntries
	tab         viewDat
	secretsPath string
}

//this doesnt work for some reason. will prob delete soon
/*
func readFromFile(m *model) (n int) {

	pstEntries, err := os.ReadFile((m.homeDir + "/.secrets"))

	if err != nil {
		if (errors.Is(err, os.ErrNotExist)) || (len(pstEntries) == 0) {

			m.data = []jsonEntries{}
			return
		}
		m.errMsg = err
	}
	m.data, err = Decrypt([]byte(m.pswdUnhashed), pstEntries)
	if err != nil {
		m.errMsg = err
	}
	return len(m.data)

}
*/

func initialModel() model {

	//initialize style

	//initalize list!

	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 20

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return model{
			errMsg: err,
		}
	}
	file, err := os.Open((homeDir + "/.jcli.json"))

	if errors.Is(err, os.ErrNotExist) { //if file doesn't exist, we know there's no password, so just setting up user to enter their new password

		os.Create((homeDir + "/.jcli.json"))
		ti.Placeholder = "enter new password"
		ti.Focus()

		m := model{

			textInput:    ti,
			pswdHash:     "",
			pswdSet:      false,
			pswdUnhashed: "",
			pswdEntered:  false,
			errMsg:       nil,
			action:       0,
			homeDir:      homeDir,
			secretsPath:  homeDir + "/.secrets",
		}

		return m

	} else if err != nil {
		return model{
			errMsg: err,
		}

	} else { // the file exists, so we hash the password and try to get them to match!
		data, err := io.ReadAll(file)
		if err != nil {
			return model{
				errMsg: err,
			}
		}
		var config conf
		err = json.Unmarshal(data, &config)
		if err != nil {
			return model{
				errMsg: err,
			}
		}
		ti.Placeholder = "enter password"
		ti.Focus()
		m := model{
			textInput: ti,

			pswdHash:     config.JournalHash,
			pswdSet:      true,
			pswdUnhashed: "",
			pswdEntered:  false,
			config:       config,
			errMsg:       nil,
			action:       0,
			homeDir:      homeDir,
			secretsPath:  homeDir + "/.secrets",
		}

		return m

	}
}

func (m model) Init() tea.Cmd {

	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//writing the part where password hasn't been entered yet
	var cmd tea.Cmd = nil
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		//general commands
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			if m.action == 1 {
				if m.list.cursor > 0 {
					m.list.cursor--
				}
			}
		case tea.KeyDown:
			if m.action == 1 {
				if m.list.cursor < len(m.list.choices) {
					m.list.cursor++
				}
			}
		case tea.KeyEsc:
			if m.action == 2 || m.action == 3 {
				m.action = 1
			}

		case tea.KeyCtrlS:
			if m.action == 2 {

				//load in data. decrypt it. add most recent entry. encrypt it. put it back

				//decrypting part!
				//since this returns nothing if the file is empty or doesn't exist, we don't have to worry about other error handling
				tmp, err := takeOutData(m.pswdUnhashed, m.secretsPath)
				if err != nil {
					m.errMsg = err
				}
				pastEntries := append(tmp, jsonEntries{Msg: m.entry.textarea.Value(), Date: time.Now()})

				//add past entries for viewing
				m.data = pastEntries
				//now must reencrypt
				err = putInFile(pastEntries, m.pswdUnhashed, m.secretsPath)
				if err != nil {
					m.errMsg = err
				}
				m.action = 1
			}

		case tea.KeyEnter:
			//all the stuff that can happen when enter is clicked!!!!!

			//password segment - if password still isn't input
			if !m.pswdEntered {

				first := sha256.New()
				if !m.pswdSet {
					//hashing what was just entered and putting it in file
					hash, err := hash(m.textInput.Value())
					if err != nil {
						m.errMsg = err
					}
					m.pswdEntered = true
					m.pswdHash = hash

					//now putting that into the file
					homeDir, err := os.UserHomeDir()

					if err != nil {
						m.errMsg = err
						return m, tea.Quit
					}

					os.WriteFile((homeDir + "/.jcli.json"), []byte("{\"JournalHash\":\""+hash+"\"}"), 0644)
					m.pswdHash = hash
					m.pswdUnhashed = m.textInput.Value()
					m.textInput.Reset()
					m.textInput.Focus()
					first.Reset()
				} else {
					hash, err := hash(m.textInput.Value())
					if err != nil {
						m.errMsg = err
					}
					if hash != m.pswdHash {
						m.pswdWrong = true
						m.textInput.Reset()
						m.textInput.Focus()
					} else {
						m.pswdEntered = true
						m.pswdUnhashed = m.textInput.Value()
						m.action = 1

						//now we have to prepare the list!
						m.list.choices = []string{"write entries", "read entries", "change password", "look at analytics", "settings", "logout"}

						m.list.cursor = 0

						return m, cmd

					}
					first.Reset()

				}

			}

			//list part
			if m.action == 1 {
				m.action = m.list.cursor + 2

				//setting up each model for when the action is clicked
				if m.action == 2 { //writing a new entry!
					m.entry.textarea = textarea.New()
					m.entry.textarea.Placeholder = "write a new entry here!"
					m.entry.textarea.Focus()
					return m, cmd
				}

			}
			//set up table here!
			if m.action == 3 {
				var rows []table.Row
				columns := []table.Column{{Title: "date written", Width: 50}}
				//if data hasn't been decrypted yet (if no entry has been written)

				if newData, err := takeOutData(m.pswdUnhashed, m.secretsPath); len(newData) == 0 { //if no data available
					if err != nil {
						m.errMsg = err
					}
					rows = []table.Row{{"no entries yet!"}}

				} else {
					m.data = newData
					rows = make([]table.Row, len(newData))

					for index, obj := range newData {
						rows[index] = table.Row{obj.Date.Format(time.RFC822)}
					}
				}

				m.tab.table = table.New(
					table.WithColumns(columns),
					table.WithRows(rows),
					table.WithFocused(true),
					table.WithHeight(7),
				)

				//rot styling copied from docs
				s := table.DefaultStyles()
				s.Header = s.Header.
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color("240")).
					BorderBottom(true).
					Bold(false)
				s.Selected = s.Selected.
					Foreground(lipgloss.Color("229")).
					Background(lipgloss.Color("57")).
					Bold(false)
				m.tab.table.SetStyles(s)

			}

		}

	default:

		if m.action == 2 {
			if !m.entry.textarea.Focused() {
				cmd = m.entry.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

		if m.action == 7 {
			return m, tea.Quit
		}
	}
	//outside tea.msg here
	if m.action == 0 {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	if m.action == 2 {
		m.entry.textarea, cmd = m.entry.textarea.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	if m.action == 3 {
		m.tab.table, cmd = m.tab.table.Update(msg)
		return m, cmd
	}

	return m, cmd
}

func (m picking) list() string {
	var s string
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	return s

}

func (m model) View() string {
	var fin string
	//just some config stuff
	if m.errMsg != nil {
		return m.errMsg.Error()
	}
	if m.debug != "" {
		return m.debug
	}
	//password segment!!!!!!!!!!!
	if !m.pswdEntered {
		var header string
		if !m.pswdSet {
			header = "welcome! a password wasn't found in this directory, so enter in a new one!"
		} else if m.pswdWrong && (m.textInput.Value() == "") {
			header = "Wrong password! Try again"
		} else {
			m.pswdWrong = false
			header = "Enter in password:"
		}

		fin = header + "\n" + m.textInput.View()
		return docStyle.Render(fin)
	}

	//password is entered here -> time to get into the actual app!

	if m.action == 1 {

		return docStyle.Render(
			"what would you like to do? \n",
			m.list.list(),
		)
	}

	if m.action == 2 {
		return docStyle.Render(
			"write entry here! \n",
			m.entry.textarea.View(),
			"\n esc to go back, ctrl + c to quit",
		)
	}

	if m.action == 3 {
		return docStyle.Render(m.tab.table.View())
	}
	//writing list part

	//never supposed to end up here
	return fmt.Sprintf("oops...", m.action)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
