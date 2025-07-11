package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().
	Margin(1, 2).
	PaddingTop(2).
	PaddingLeft(4).
	Align(lipgloss.Left)

// constant for formatting time
var timeFormat = "Mon Jan 2 3:04pm"

// will take in a password and store the hash in a config file - if no such variable exists, offer to make a new one. the password will
// be the key to decrypt the file with the journal entries. a password is only required to read the past entries, not current
type conf struct {
	JournalHash string `json:"JournalHash"`
}

// setting up the list part

// typing entries part

type entry struct {
	Title string    `json:"Title"`
	Msg   string    `json:"Msg"`
	Date  time.Time `json:"Date"`
	Tags  []string  `json:"Tags"`
}
type jsonEntries struct { //json struct for single entry
	readIn  int            //initialized to zero, when read in and empty, set to 1 so that we dont have to keep rereading and returning nothing
	Entries []entry        `json:"entries"` //will implement tags tmrw! (today)
	Tags    map[string]int `json:"tags"`    //all UNIQUE tags
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
	action  int         //what r u doing rn?
	data    jsonEntries //LARGE json object in here
	loading bool        //for implementing loading mechanism
	saving  bool        //so that program doesn't quit before saving is finished
	//initial list used to select action
	list picking

	//storing data

	tab         viewDat
	secretsPath string

	//input
	entryView writing

	//ui
	width  int
	height int
}
type loading bool

func setLoading() tea.Msg {
	return loading(true)
}

func initialModel(args []string) model {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return model{
			errMsg: err,
		}
	}

	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 20

	//so that action cna be pre-set when testing so that password doesn't have to be entered
	if len(os.Args) > 1 { //will HAVE to delete this later just bypasses whole password part

		pswdHash, _ := hash("password")
		action, _ := strconv.Atoi(args[1])
		m := model{

			textInput:    ti,
			pswdHash:     pswdHash,
			pswdSet:      true,
			pswdUnhashed: "password",
			pswdEntered:  true,
			errMsg:       nil,
			action:       action,
			homeDir:      homeDir,
			secretsPath:  homeDir + "/.secrets",
			//really just a hack
			list: picking{choices: []string{"write entries", "read entries", "change password", "look at analytics", "settings", "logout"}, cursor: 0},
		}
		if action == 3 {
			m.readInit()
		}
		return m

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

	//special cases
	switch msg := msg.(type) {
	case dataLoadedIn:
		m.data = msg.data
		m.saving = false
		m.loading = false
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:

		switch msg.Type {
		//general commands
		case tea.KeyCtrlC:
			if !m.saving {
				return m, tea.Quit
			}

		case tea.KeyEsc:
			m.action = 1

		}

	}

	//if no special cases, -> just pass off to helper update functions
	if m.action == 1 {
		return m.listUpdate(msg)
	}

	if m.action == 2 {
		return m.writingUpdate(msg)
	}

	if m.action == 3 {
		return m.readUpdate(msg)

	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		//general commands

		case tea.KeyEsc:
			if m.action == 2 || m.action == 3 || m.action == 5 {
				m.action = 1
			}

		case tea.KeyCtrlS:
			//saved new password
			if m.action == 4 {
				//have to take old data, reencrypt it with new password, put it back,
				// then rehash password and put that back too
				newPswd := m.textInput.Value()
				if data, err := takeOutData(m.pswdUnhashed, m.homeDir); len(data.Entries) != 0 {
					if err != nil {
						m.errMsg = err
						return m, nil
					}
					//we know there's data, now we have to reset the password
					err = putInFile(data, newPswd, m.secretsPath)
					if err != nil {
						m.errMsg = err
						return m, nil
					}
				}
				//now writing pswd hash into file
				newHash, err := hash(newPswd)
				m.pswdHash = newHash
				if err != nil {
					m.errMsg = err
					return m, nil
				}
				//TODO when more config options are added in, this will have to load in the config file first then change it
				//to avoid overwriting the rest of the config
				os.WriteFile((m.homeDir + "/.jcli.json"), []byte("{\"JournalHash\":\""+newHash+"\"}"), 0644)
				m.pswdUnhashed = newPswd

				//now that that's all done, just return back to user
				m.textInput.Reset()

				m.action = 1
				return m.listInit()
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

					if err != nil {
						m.errMsg = err
						return m, tea.Quit
					}

					os.WriteFile((m.homeDir + "/.jcli.json"), []byte("{\"JournalHash\":\""+hash+"\"}"), 0644)
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
						//password is correct!
						m.pswdEntered = true
						m.pswdUnhashed = m.textInput.Value()
						m.action = 1
						m.textInput.Reset()

						//now we have to prepare the list!
						return m.listInit()

					}
					first.Reset()

				}

			}

			//list part

			//set up table here!
			if m.action == 3 {
				m.readInit()
			}

		} //here is where tea.enter ends

	default:

		if m.action == 7 {
			return m, tea.Quit
		}
	}
	//outside tea.msg here
	if m.action == 0 || m.action == 4 { //the two options with one line inputs
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
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

		return m.listView()
	}

	if m.action == 2 {
		return m.writingView()
	}

	if m.action == 3 {
		return m.readView()
	}

	//resetting password
	if m.action == 4 {
		return docStyle.Render(
			"write new password here: \n",
			m.textInput.View(),
			"\n esc to go back, ctrl+s to save",
		)
	}

	if m.action == 5 {
		return m.viewAggs()
	}

	//never supposed to end up here
	return fmt.Sprintf("oops...", m.action)
}

//flags!

func main() {

	p := tea.NewProgram(initialModel(os.Args))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
