package main

import (
	"fmt"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

type GenerateCommand struct {
	rawCode     string
	onlyChannel uint
	onlyAction  string
	verbose     bool
	tabulate    bool
}

func configureGenerateCommand(app *kingpin.Application) {
	// Given a single 'seed' capture, we can generate 63 additional remote codes.
	c := &GenerateCommand{}
	generate := app.Command("generate", "Generate additional remote codes given a single seed code").Action(c.Run)
	generate.Arg("code", "RF code to use as a seed").Required().StringVar(&c.rawCode)
	generate.Flag("only-channel", "Generate codes only for this channel").Default("255").UintVar(&c.onlyChannel)
	generate.Flag("only-action", "Generate codes only for this action").
		EnumVar(&c.onlyAction, "DOWN", "STOP", "UP", "PAIR")
	generate.Flag("verbose", "Output additional information per-line").BoolVar(&c.verbose)
	generate.Flag("tabulate", "Tabulate output data").BoolVar(&c.tabulate)
}

func (g *GenerateCommand) printCodeV(code *RemoteCode) {
	if g.verbose {
		fmt.Printf("%s     C: %2d   A: %4s (%d)\n", code.Serialize(), code.Channel, code.Action.Name, code.Action.Value)
	} else if g.tabulate {
		fmt.Printf("%d,\"%s\",%d,%s\n", code.Channel, code.Action.Name, code.Action.Value, strings.Join(strings.Split(code.Serialize(), ""), ","))
	} else {
		fmt.Println(code.Serialize())
	}

}

func (g *GenerateCommand) generateActions(code *RemoteCode) error {
	if g.onlyAction != "" {
		action, err := ActionFromName(g.onlyAction)
		if err != nil {
			return err
		}
		code.SetAction(action)
		g.printCodeV(code)
	} else {
		// Generate for all actions
		for _, actionName := range []string{"UP", "STOP", "DOWN", "PAIR"} {
			action, _ := ActionFromName(actionName)
			code.SetAction(action)
			g.printCodeV(code)
		}
	}
	return nil
}

func (g *GenerateCommand) Run(c *kingpin.ParseContext) error {
	// Let's make things simple - get the base code for this channel.
	code, err := Deserialize(g.rawCode)
	if err != nil {
		return err
	}

	if g.onlyChannel != 255 {
		code.SetChannel(g.onlyChannel)
		err := g.generateActions(code)
		if err != nil {
			return err
		}
	} else {
		// Generate for all channels.
		for i := 0; i < 64; i++ {
			code.SetChannel(uint(i))
			err := g.generateActions(code)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
