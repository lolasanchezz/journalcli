package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// constant for formatting time
var timeFormat = "Mon Jan 2 3:04pm"

// will take in a password and store the hash in a config file - if no such variable exists, offer to make a new one. the password will
// be the key to decrypt the file with the journal entries. a password is only required to read the past entries, not current

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
	width   int
	height  int
	aWidth  int //actual width
	aHeight int //actual height
	help    help.Model

	styles styles
	//pswd reset
	psRs pswdReset

	//settings
	settings settingInp

	//destroy everything!!
	erase erase

	//aggsss
	aggs aggs
}
type loading bool

func setLoading() tea.Msg {
	return loading(true)
}

func initialModel() model {

	//make the default styles
	defStyles := styles{root: rootStyle, viewport: viewportStyle, header: headerStyle, filter: searchBoxStyle}

	homeDir, err := os.UserHomeDir()
	confPath := homeDir + "/.jcli.json"
	if err != nil {
		return model{
			errMsg: err,
		}
	}

	ti := textinput.New()
	ti.CharLimit = 156

	file, err := os.Open(confPath)

	if errors.Is(err, os.ErrNotExist) { //if file doesn't exist, we know there's no password, so just setting up user to enter their new password

		ti.Placeholder = "enter new password"
		ti.Width = lipgloss.Width(ti.Placeholder)
		ti.Focus()

		m := model{
			help:         help.New(),
			confPath:     confPath,
			pswdInput:    pswdEnter{ti: ti, pswdSet: false, pswdEntered: false},
			pswdHash:     "",
			config:       defaultStyles,
			pswdUnhashed: "",
			styles:       defStyles,
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
		//set up the styling real QUICK

		defStyles.root = defStyles.root.Foreground(lipgloss.Color(config.TextColor)).BorderForeground(lipgloss.Color(config.BordCol))
		defStyles.header = defStyles.header.Foreground(lipgloss.Color(config.SecTextColor))
		defStyles.filter = defStyles.filter.BorderForeground(lipgloss.Color(config.BordCol)).Foreground(lipgloss.Color(config.TextColor))

		ti.Placeholder = "enter password"
		ti.Focus()
		ti.Width = lipgloss.Width(ti.Placeholder)
		m := model{
			help:         help.New(),
			pswdInput:    pswdEnter{ti: ti, pswdSet: true, pswdEntered: false},
			confPath:     confPath,
			pswdHash:     config.JournalHash,
			styles:       defStyles,
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
	if m.config.Fullscreen {

		return tea.Batch(textinput.Blink, tea.EnterAltScreen)
	}
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
		if m.config.Fullscreen {
			m.aWidth = msg.Width - 2
			m.aHeight = msg.Height - 2
		} else {
			m.aWidth = int(float64(msg.Width) * m.config.Width)
			m.aHeight = int(float64(msg.Height) * m.config.Height)
		}
		m.styles.root = m.styles.root.Width(m.aWidth).Height(m.aHeight)
		m.help.Width = m.aWidth

	case tea.KeyMsg:

		switch msg.Type {
		//general commands
		case tea.KeyCtrlC:
			if !m.saving {
				return m, tea.Quit
			}
		case tea.KeyEsc:
			if !(m.action == 0 || m.action == 7) {
				m.action = 1
			}
		}
	}
	switch m.action {
	//if no special cases, -> just pass off to helper update functions
	case 0:
		return m.pswdUpdate(msg)

	case 1:
		return m.listUpdate(msg)

	case 2:
		return m.writingUpdate(msg)

	case 3:
		return m.readUpdate(msg)

	case 4:
		return m.psrsUpdate(msg)

	case 6:
		return m.settingsUpdate(msg)

	case 7:
		return m, tea.Quit

	case 8:
		return m.eraseUpdate(msg)

	}
	return m, cmd
}

func (m model) View() string {
	var str string

	//just some config stuff
	if m.errMsg != nil {
		return m.errMsg.Error()
	}
	if m.debug != "" {
		return m.debug
	}
	//password segment!!!!!!!!!!!

	//password is entered here -> time to get into the actual app!
	switch m.action {
	case 0:
		str = m.pswdView()

	case 1:

		str = m.listView()
	case 2:
		str = m.writingView()

	case 3:
		str = m.readView()

	//resetting password
	case 4:
		str = m.psrsView()

	case 5:
		str = m.aggsView()

	case 6:
		str = m.settingsView()
	case 8:
		str = m.eraseView()
	default:
		return ("something went wrong." + strconv.Itoa(m.action))
	}

	return m.styles.root.Render(m.addHelp(str))
	//never supposed to end up here

}

///flags!

func main() {

	p := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())
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
	} else if arr, ok := v.(table.Row); ok {
		var s string
		for _, val := range arr {
			s += val + " "
		}
		d = []byte(s)
	} else {
		d, err = json.Marshal(v)
		if err != nil {
			d = []byte(err.Error())
		}
	}

	_ = os.WriteFile("./debug.txt", d, 0644)
}
