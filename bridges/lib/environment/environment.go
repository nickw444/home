package environment

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

type MQTT struct {
	Broker        string
	User          string
	Password      string
	TlsSkipVerify bool
	Dummy         bool
}

type HC struct {
	AccessCode string
	Port       string
	Debug      bool
}

type Environment struct {
	app        *kingpin.Application
	Debug      bool
	ConfigFile string
	Mqtt       *MQTT
	HC         *HC
}

func New(app *kingpin.Application) *Environment {
	return &Environment{
		app: app,
	}
}

func (e *Environment) WithConfigFile(defaultFilename string) *Environment {
	e.app.Flag("config", "Provide a configuration file.").Short('c').Default(defaultFilename).StringVar(&e.ConfigFile)
	return e
}

func (e *Environment) WithDebug() *Environment {
	e.app.Flag("debug", "Enable Debug output").BoolVar(&e.Debug)
	return e
}

func (e *Environment) WithHC() *Environment {
	hc := &HC{}
	e.app.Arg("access-code", "Homekit Access code to use").Required().StringVar(&hc.AccessCode)
	e.app.Flag("port", "Port for Homekit to listen on.").Short('p').StringVar(&hc.Port)
	e.app.Flag("hc-debug", "Enable Debug output for HC library").BoolVar(&hc.Debug)
	e.HC = hc
	return e
}

func (e *Environment) WithMQTT() *Environment {
	mqtt := &MQTT{}
	e.app.Flag("dummy-mqtt", "Use a dummy MQTT instance").BoolVar(&mqtt.Dummy)
	e.app.Flag("mqtt-broker", "MQTT Broker URL").Default("tls://127.0.0.1:8883").StringVar(&mqtt.Broker)
	e.app.Flag("mqtt-user", "MQTT Broker User").StringVar(&mqtt.User)
	e.app.Flag("mqtt-password", "MQTT Password").StringVar(&mqtt.Password)
	e.app.Flag("mqtt-tls-skip-verify", "Skip TLS cert verification").BoolVar(&mqtt.TlsSkipVerify)
	e.Mqtt = mqtt
	return e
}

func (e *Environment) Parse() *Environment {
	kingpin.MustParse(e.app.Parse(os.Args[1:]))
	return e
}
