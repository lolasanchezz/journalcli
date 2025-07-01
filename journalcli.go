package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// will take in a password and store the hash in a config file - if no such variable exists, offer to make a new one. the password will
// be the key to decrypt the file with the journal entries. a password is only required to read the past entries, not current.
type conf struct {
	JournalHash string `json:"JournalHash"`
}
type model struct {
	//entering password part

	pswdSet      bool   // does password exist in .env?
	pswdEntered  bool   // has user entered in correct password?
	pswdHash     string // password in .env
	pswdUnhashed string // correct password entered in by user (real password)
	pswdWrong    bool   // just showing whether entered in password is incorrect (temporary flash "wrong! in header")
	errMsg       error  //for passing along errors to stdout
	textInput    textinput.Model
	config       conf
	debug        string
}

func initialModel() model {

	//initaliz

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

		return model{
			textInput:    ti,
			pswdHash:     "",
			pswdSet:      false,
			pswdUnhashed: "",
			pswdEntered:  false,
			errMsg:       nil,
		}

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
		return model{
			textInput: ti,

			pswdHash:     config.JournalHash,
			pswdSet:      true,
			pswdUnhashed: "",
			pswdEntered:  false,
			config:       config,
			errMsg:       nil,
		}
	}
}

func (m model) Init() tea.Cmd {

	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//writing the part where password hasn't been entered yet
	var cmd tea.Cmd = nil

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		//all the stuff that can happen when enter is clicked!!!!!
		case tea.KeyEnter:

			//password segment - if password still isn't input
			if !m.pswdEntered {
				first := sha256.New()
				if !m.pswdSet { //hashing what was just entered and putting it in file
					_, err := first.Write([]byte(m.textInput.Value()))
					if err != nil {
						m.errMsg = err
					}

					hash := first.Sum(nil)
					strHash := hex.EncodeToString(hash[:])

					if err != nil {
						m.errMsg = err
					}
					m.pswdEntered = true
					m.pswdHash = strHash

					//now putting that into the file
					homeDir, err := os.UserHomeDir()

					if err != nil {
						m.errMsg = err
						return m, tea.Quit
					}

					os.WriteFile((homeDir + "/.jcli.json"), []byte("{\"JournalHash\":\""+strHash+"\"}"), 0644)
					m.pswdHash = strHash
					m.pswdUnhashed = m.textInput.Value()
					m.textInput.Reset()
					m.textInput.Focus()
					first.Reset()
				} else {
					_, err := first.Write([]byte(m.textInput.Value()))
					if err != nil {
						m.errMsg = err
					}
					hash := first.Sum(nil)

					strHash := hex.EncodeToString(hash[:])

					if strHash != m.pswdHash {
						m.pswdWrong = true
						m.textInput.Reset()
						m.textInput.Focus()
					} else {
						m.pswdEntered = true
						m.pswdUnhashed = m.textInput.Value()

					}
					first.Reset()

				}
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m model) View() string {
	if m.errMsg != nil {
		return m.errMsg.Error()
	}
	if m.debug != "" {
		return m.debug
	}
	if !m.pswdEntered {
		var header string
		if !m.pswdSet {
			header = "welcome! a password wasn't found in this directory, so enter in a new one!"
		} else if m.pswdWrong && (m.textInput.Value() == "") {
			header = "Wrong password! Try again. the hash of what was entered is: \n" + m.pswdHash
		} else {
			m.pswdWrong = false
			header = "Enter in password:"
		}

		return header + "\n" + m.textInput.View()
	}

	//password is entered here -> time to get into the actual app!

	return "heres the hash!" + m.pswdHash

}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
