package main

import (
	"time"

	"regexp"

	"strings"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/nickw444/homekit/bridges/lib/environment"
	"github.com/nickw444/homekit/bridges/lib/mqtt"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var log = logrus.New()

var unsafeSerialChars = regexp.MustCompile("[^a-z0-9\\-]")

func main() {
	app := kingpin.New("Blindkit", "Homekit Blind Bridge")
	env := environment.New(app).
		WithConfigFile("blindkit.conf.yml").
		WithDebug().
		WithHC().
		WithMQTT().
		Parse()

	if env.Debug {
		log.Level = logrus.DebugLevel
	}

	mqttClient := GetMQTTClient(env.Mqtt)
	config := parseConfig(env.ConfigFile)

	// Create the bridge.
	bridge := accessory.New(accessory.Info{
		Name:         config.Name,
		Manufacturer: config.Manufacturer,
		Model:        "blindkit",
	}, accessory.TypeBridge)

	// Get the transmitters
	rfServices := make(map[string]*RFService)
	for _, rfConfig := range config.Transmitters {
		rfServices[rfConfig.ID] = NewRFService(rfConfig.Serial, mqttClient)
	}

	// Make the blinds
	var accessories []*accessory.Accessory
	for _, blindConfig := range config.Blinds {
		rfService := rfServices[blindConfig.Transmitter]
		if rfService == nil {
			log.Panic("Unknown RF Transmitter", blindConfig.Transmitter)
		}
		client := NewRaexBlindClient(blindConfig.Remote, blindConfig.Channel,
			3, time.Second, rfService)
		serial := unsafeSerialChars.ReplaceAllString(strings.ToLower(blindConfig.Name), "-")
		log := log.WithField("serial", serial)
		blind := NewRaexBlind(client, log, serial, blindConfig.Name)
		accessories = append(accessories, blind.accessory)
		log.Infof("Registered blind!")
	}

	t, err := hc.NewIPTransport(hc.Config{Pin: env.HC.AccessCode, Port: env.HC.Port}, bridge, accessories...)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	log.Printf("Started server with access code: '%s' on port %s", env.HC.AccessCode, env.HC.Port)
	t.Start()
}

func GetMQTTClient(conf *environment.MQTT) (c mqtt.Client) {
	if conf.Dummy {
		return &mqtt.LoggingClient{
			Log: log.WithField("component", "logging-mqtt-client"),
		}
	} else {
		logger := log.WithField("component", "paho-mqtt-client")
		client := mqtt.NewPahoClient(conf.Broker, conf.User, conf.Password, conf.TlsSkipVerify, logger)
		if err := client.Connect(); err != nil {
			log.Panic(err)
		}
		return client
	}
}
