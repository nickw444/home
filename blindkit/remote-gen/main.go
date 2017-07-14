package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("RAEX Remote Code Generator Tool", "")
	configureInfoCommand(app)
	configureCreateCommand(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
