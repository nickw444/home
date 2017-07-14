package main

import (
	"fmt"

	"github.com/nickw444/homekit/blindkit/remote-gen/remote_code"
	"gopkg.in/alecthomas/kingpin.v2"
)

type InfoCommand struct {
	rawCode  string
	validate bool
}

func configureInfoCommand(app *kingpin.Application) {
	c := &InfoCommand{}
	identify := app.Command("info", "Extract information about a captured remote code").Action(c.Run)
	identify.Arg("code", "RF code to identify").Required().StringVar(&c.rawCode)
	identify.Flag("validate", "Verify that the checksum matches our guessed checksum value.").BoolVar(&c.validate)
}

func (i *InfoCommand) Run(c *kingpin.ParseContext) error {
	fmt.Println("Raw Code:  ", i.rawCode)
	code, err := remote_code.Deserialize(i.rawCode)
	if err != nil {
		return err
	}

	fmt.Println("Channel:   ", code.Channel)
	fmt.Println("Remote:    ", code.Remote)
	fmt.Println("Action:    ", code.Action.Name)
	fmt.Println("Checksum:         ", code.Checksum)
	fmt.Println("Guessed Checksum: ", code.GuessChecksum())

	if i.validate {
		if code.Checksum != code.GuessChecksum() {
			return fmt.Errorf("The guessed checksum value does not match the provided one. %d != %d", code.Checksum, code.GuessChecksum())
		}
		if i.rawCode != code.Serialize() {
			return fmt.Errorf("The serialized code is not equal to the input code. %s\n%s", i.rawCode, code.Serialize())
		}
	}

	return nil
}
