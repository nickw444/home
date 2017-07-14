package main

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type CreateCommand struct {
	channel uint8
	remote  uint16
	actions []string
	verbose bool
}

func configureCreateCommand(app *kingpin.Application) {
	c := &CreateCommand{}
	create := app.Command("create", "Create new remote codes").Action(c.Run)
	create.Flag("channel", "Channel to broadcast on (uint8)").Required().Uint8Var(&c.channel)
	create.Flag("remote", "Remote value to broadcast (uint16)").Required().Uint16Var(&c.remote)
	create.Flag("action", "Actions to create").EnumsVar(&c.actions, GetValidNames()...)

	create.Flag("verbose", "Output additional information about the generated codes").BoolVar(&c.verbose)
}

func (c *CreateCommand) Run(p *kingpin.ParseContext) error {
	if len(c.actions) == 0 {
		c.actions = GetValidNames()
	}

	code := RemoteCode{
		LeadingBit: 0,
		Channel:    c.channel,
		Remote:     RemoteValue(c.remote),
	}

	if err := code.Check(); err != nil {
		return err
	}

	for _, actionName := range c.actions {
		code.Action, _ = ActionFromName(actionName)
		code.Checksum = code.GuessChecksum()
		if c.verbose {
			fmt.Printf("%s    Action: %-4s\n", code.Serialize(), code.Action.Name)
		} else {
			fmt.Println(code.Serialize())
		}

	}

	return nil
}
