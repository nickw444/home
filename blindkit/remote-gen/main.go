package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("Remote Generator", "")
	configureGenerateCommand(app)
	configureInfoCommand(app)
	configureCreateCommand(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// fmt.Println("00000010011110100110010111011111101000100")
	// code, err := Deserialize("00000010011110100110010111011111101000100")
	// if err != nil {
	// 	panic(err)
	// }
	// // fmt.Println(code)
	// fmt.Println(code.Serialize())
}
