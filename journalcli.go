package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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

	pswdHash     string // password in .env
	pswdUnhashed string // correct password entered in by user (real password)

	errMsg    error
	pswdInput pswdEnter //for passing along errors to stdout

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
	confPath    string
	//input
	entryView writing

	//ui
	width  int
	height int

	//pswd reset
	psRs pswdReset
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
	confPath := homeDir + "/.jcli.json"
	file, err := os.Open(confPath)

	if errors.Is(err, os.ErrNotExist) { //if file doesn't exist, we know there's no password, so just setting up user to enter their new password

		os.Create(confPath)
		ti.Placeholder = "enter new password"
		ti.Focus()

		m := model{
			confPath:  confPath,
			pswdInput: pswdEnter{ti: ti, pswdSet: false, pswdEntered: false},
			pswdHash:  "",

			pswdUnhashed: "",

			errMsg:      nil,
			action:      0,
			homeDir:     homeDir,
			secretsPath: homeDir + "/.secrets",
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

			pswdInput: pswdEnter{ti: ti, pswdSet: true, pswdEntered: false},
			confPath:  confPath,
			pswdHash:  config.JournalHash,

			pswdUnhashed: "",

			config:      config,
			errMsg:      nil,
			action:      0,
			homeDir:     homeDir,
			secretsPath: homeDir + "/.secrets",
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
		rootStyle.Width(msg.Width)
		rootStyle.Height(msg.Height / 3)
		return m, nil

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
	if m.action == 0 {
		return m.pswdUpdate(msg)
	}
	if m.action == 1 {
		return m.listUpdate(msg)
	}

	if m.action == 2 {
		return m.writingUpdate(msg)
	}

	if m.action == 3 {
		return m.readUpdate(msg)

	}

	if m.action == 4 {
		return m.psrsUpdate(msg)
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		//general commands

		case tea.KeyEsc:
			if m.action == 2 || m.action == 3 || m.action == 5 {
				m.action = 1
			}
			/*
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
							m.action = 1
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
			*/

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

	return m, cmd
}

func (m model) View() string {

	//just some config stuff
	if m.errMsg != nil {
		return m.errMsg.Error()
	}
	if m.debug != "" {
		return m.debug
	}
	//password segment!!!!!!!!!!!

	//password is entered here -> time to get into the actual app!
	if m.action == 0 {
		return rootStyle.Render(m.pswdView())
	}
	if m.action == 1 {

		return rootStyle.Render(m.listView())
	}

	if m.action == 2 {
		return rootStyle.Render(m.writingView())
	}

	if m.action == 3 {
		return rootStyle.Render(m.readView())
	}

	//resetting password
	if m.action == 4 {
		return rootStyle.Render(m.psrsView())
	}

	if m.action == 5 {
		return rootStyle.Render(m.viewAggs())
	}

	//never supposed to end up here
	return fmt.Sprintf("oops...", m.action)
}

///flags!

func main() {

	p := tea.NewProgram(initialModel(os.Args))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func debug(v any) {
	var d []byte
	if err, ok := v.(error); ok {
		str := err.Error()
		d = []byte(str)
	} else if str, ok := v.(string); ok {
		d = []byte(str)
	} else {
		return
	}
	os.WriteFile("./debug.txt", d, os.FileMode(os.O_RDWR))
}
