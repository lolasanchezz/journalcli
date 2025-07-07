package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/scrypt"
)

// hashing function
func hash(val string) (string, error) {
	first := sha256.New()
	_, err := first.Write([]byte(val))
	if err != nil {
		return "", err
	}

	hash := first.Sum(nil)
	strHash := hex.EncodeToString(hash[:])

	return strHash, nil
}

// aes needs a max 32 byte key and password won't necessarily be that, so this generates such a key
// with scrypt. will probably end up using this same function for the hash
func getKey(password, salt []byte) ([]byte, []byte, error) {
	//if salt wasn't passed in, make a new one!
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 1048576, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil

}

// encrypting to put into file
func Encrypt(pswd, data []byte) ([]byte, error) {
	key, salt, err := getKey(pswd, nil)
	if err != nil {
		return nil, err
	}
	ciph, err := aes.NewCipher(key) //the key used to encrypt stuff!
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(ciph) //wrapping the key in a interface that allows me to encrypt all my data at once
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())     //an empty slice that has enough space for the nonce needed to decrypt the data
	if _, err = rand.Read(nonce); err != nil { //making the actual nonce!
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil) //yay!! the final thing
	//add salt too!
	ciphertext = append(ciphertext, salt...)
	return ciphertext, nil
}

// also turns into json object
func Decrypt(key, data []byte) (jsonEntries, error) {

	salt, data := data[len(data)-32:], data[:len(data)-32]

	key, _, err := getKey(key, salt)
	if err != nil {
		return jsonEntries{}, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return jsonEntries{}, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return jsonEntries{}, err
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return jsonEntries{}, err
	}

	//now decrypting json
	var entries jsonEntries
	err = json.Unmarshal(plaintext, &entries)
	if err != nil {
		return jsonEntries{}, err
	}
	return entries, nil
}

func putInFile(data jsonEntries, password string, path string) error {
	mshData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	cipherText, err := Encrypt([]byte(password), mshData)
	if err != nil {
		return err
	}

	os.WriteFile(path, cipherText, 0644)
	return nil
}

// take in only password, return []jsonEntries
func takeOutData(password string, path string) (jsonEntries, error) {

	newData, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return jsonEntries{readIn: 1}, nil
		}
		return jsonEntries{}, err
	}

	decData, err := Decrypt([]byte(password), newData)
	if err != nil {
		log.Fatal(err)
	}
	decData.readIn = 1 //read in successfully
	return decData, nil

}

type errMsg struct {
	err error
}

func takeOutDataCmd(password string, path string) tea.Cmd {
	return func() tea.Msg {
		data, err := takeOutData(password, path)
		if err != nil {
			return errMsg{err} // Return an error message if something goes wrong
		}
		return data
	}
}

func putInFileCmd(data jsonEntries, password string, path string) tea.Cmd {
	return func() tea.Msg {
		if err := putInFile(data, password, path); err != nil {
			return errMsg{err: err}
		}
		return errMsg{err: nil}
	}
}
