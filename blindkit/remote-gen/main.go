package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var patternFmt = "0{channel}{remote}011{action}11111{action'}{channel'}{m'}{action''}"

type IdentifyCommand struct {
	rawCode string
}

func configureIdentifyCommand(app *kingpin.Application) {
	c := &IdentifyCommand{}
	identify := app.Command("identify", "Identify a remote code, extracting information").Action(c.Run)
	identify.Arg("code", "RF code to identify").Required().StringVar(&c.rawCode)
}

func (i *IdentifyCommand) Run(c *kingpin.ParseContext) error {
	fmt.Println("Raw Code:   ", i.rawCode)
	code, err := Deserialize(i.rawCode)
	if err != nil {
		return err
	}

	fmt.Println("Serialized: ", code.Serialize())
	fmt.Println("")

	fmt.Println("Channel:  ", code.Channel)
	fmt.Println("Remote:   ", code.Remote)
	fmt.Printf("Action:    %d (%s)", code.Action, getActionName(code.Action))
	fmt.Println("")
	fmt.Println("ChannelF: ", code.ChannelF)
	fmt.Println("MF:       ", code.MF)
	fmt.Println("ActionF:  ", code.ActionF)
	fmt.Println("ActionFF: ", code.ActionFF)

	return nil
}

type GenerateCommand struct {
	rawCode     string
	onlyChannel int
	onlyAction  string
	verbose     bool
	tabulate    bool
}

func configureGenerateCommand(app *kingpin.Application) {
	// Given a single 'seed' capture, we can generate 63 additional remote codes.
	c := &GenerateCommand{}
	generate := app.Command("generate", "Generate additional remote codes given a single seed code").Action(c.Run)
	generate.Arg("code", "RF code to use as a seed").Required().StringVar(&c.rawCode)
	generate.Flag("only-channel", "Generate codes only for this channel").Default("-1").IntVar(&c.onlyChannel)
	generate.Flag("only-action", "Generate codes only for this action").
		EnumVar(&c.onlyAction, "DOWN", "STOP", "UP", "PAIR")
	generate.Flag("verbose", "Output additional information per-line").BoolVar(&c.verbose)
	generate.Flag("tabulate", "Tabulate output data").BoolVar(&c.tabulate)
}

func (g *GenerateCommand) printCodeV(code *RemoteCode) {
	if g.verbose {
		fmt.Printf("%s     C: %2d   A: %4s (%d)\n", code.Serialize(), code.Channel, getActionName(code.Action), code.Action)
	} else if g.tabulate {
		fmt.Printf("%d,\"%s\",%d,%s\n", code.Channel, getActionName(code.Action), code.Action, strings.Join(strings.Split(code.Serialize(), ""), ","))
	} else {
		fmt.Println(code.Serialize())
	}

}

func (g *GenerateCommand) generateActions(code *RemoteCode) error {
	if g.onlyAction != "" {
		actionVal, err := getActionValue(g.onlyAction)
		if err != nil {
			return err
		}
		code.SetAction(actionVal)
		g.printCodeV(code)
	} else {
		// Generate for all actions
		for _, i := range []int{2, 1, 0, 3} {
			code.SetAction(i)
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

	if g.onlyChannel != -1 {
		code.SetChannel(g.onlyChannel)
		err := g.generateActions(code)
		if err != nil {
			return err
		}
	} else {
		// Generate for all channels.
		for i := 0; i < 64; i++ {
			code.SetChannel(i)
			err := g.generateActions(code)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	app := kingpin.New("Remote Generator", "")
	configureGenerateCommand(app)
	configureIdentifyCommand(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// fmt.Println("00000010011110100110010111011111101000100")
	// code, err := Deserialize("00000010011110100110010111011111101000100")
	// if err != nil {
	// 	panic(err)
	// }
	// // fmt.Println(code)
	// fmt.Println(code.Serialize())
}
