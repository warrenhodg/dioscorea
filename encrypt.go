package main

import (
	"errors"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var encryptCommand *kingpin.CmdClause
var encryptUsername *string
var encryptPrivateKeyFile *string

func setupEncryptCommand(app *kingpin.Application) {
	encryptCommand := app.Command("encrypt", "Encrypt the encryptable parts of the file")

	encryptUsername = encryptCommand.Flag("user", "Name of user").Short('u').Default(os.Getenv("USER")).String()
	encryptPrivateKeyFile = encryptCommand.Flag("pvt-key", "Filename of private key").Short('k').Default(os.Getenv("HOME") + "/.ssh/id_rsa").String()
}

func handleEncryptCommand(commands []string) error {
	teamPassFile, err := readFile(filename, true)
	if err != nil {
		return err
	}

	_, found := teamPassFile.Users[*encryptUsername]
	if !found {
		return errors.New(fmt.Sprintf("User not found : %s", *encryptUsername))
	}

	for groupName, group := range teamPassFile.Groups {
		encSymKey, found := group.Keys[*encryptUsername]
		if found {
			if group.Decrypted == nil {
				group.Decrypted = make(map[string]string)
			}

			symKey, err := decryptSymmetricalKey(encSymKey, *encryptPrivateKeyFile)
			if err != nil {
				return err
			}

			for valueName, value := range group.Decrypted {
				mustAdd := true

				encValue, found := group.Values[valueName]
				if found {
					decValue, err := decryptValue(symKey, encValue)
					if err != nil {
						return err
					}

					if decValue == value {
						mustAdd = false
					}
				}

				if mustAdd {
					newEncValue, err := encryptValue(symKey, value)
					if err != nil {
						return err
					}

					group.Values[valueName] = newEncValue
				}
			}

			group.Decrypted = nil

			teamPassFile.Groups[groupName] = group
		} else {
			if group.Decrypted != nil {
				return errors.New(fmt.Sprintf("There are plain-text values in a group of which you are not a part : %s", groupName))
			}
		}
	}

	return writeFile(filename, false, teamPassFile)
}