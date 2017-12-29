package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hcLog "github.com/brutella/hc/log"
	"github.com/nickw444/homekit/bridges/lib/environment"
	"github.com/nickw444/homekit/bridges/lib/mqtt"
	hca "github.com/nickw444/homekit/bridges/mainbridge/accessories"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var log = logrus.New()

func main() {
	app := kingpin.New("MQTTBridge", "Homekit MQTT Bridge")
	env := environment.New(app).
		WithConfigFile("bridge.conf.yml").
		WithDebug().
		WithHC().
		WithMQTT().
		Parse()

	if env.Debug {
		log.Level = logrus.DebugLevel
	}

	if env.HC.Debug {
		hcLog.Debug.Enable()
		hcLog.Info.Enable()
	}

	mqttClient := GetMQTTClient(env.Mqtt)
	config := parseConfig(env.ConfigFile)

	// Create the bridge.
	bridge := accessory.New(accessory.Info{
		Name:         config.Name,
		Manufacturer: config.Manufacturer,
		Model:        config.Model,
	}, accessory.TypeBridge)

	// Make the accessories
	var accessories []*accessory.Accessory
	for _, accConf := range config.Accessories {
		acc := makeAccessory(mqttClient, accConf)
		accessories = append(accessories, acc.GetHCAccessory())
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

func makeAccessory(mqttClient mqtt.Client, conf *accessoryConfig) (acc hca.HCAccessory) {
	logger := log.WithField("accessory", conf.Model).
		WithField("serial", conf.Serial)

	logger.Infof("Loading accessory...")

	if conf.Model == "sonoff-switch" {
		switchConfig := hca.NewSonoffSwitchConfig(conf.Conf)
		acc = hca.NewSonoffSwitch(switchConfig, mqttClient, conf.Serial, conf.Name)
	} else if conf.Model == "garagedoor" {
		acc = hca.NewGarageDoor(mqttClient, conf.Serial, conf.Name, logger)
	} else if conf.Model == "sonoff-th10" {
		acc = hca.NewSonoffTH10(mqttClient, conf.Serial, conf.Name)
	} else if conf.Model == "latch-lock" {
		lockConfig := hca.NewLatchLockConfig(conf.Conf)
		acc = hca.NewLatchLock(lockConfig, mqttClient, conf.Serial, conf.Name, logger)
	} else {
		log.Panicf("Not a valid accessory model: '%s'", conf.Model)
	}

	acc.Start()

	return acc
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
