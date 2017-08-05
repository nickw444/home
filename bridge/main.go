package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hcLog "github.com/brutella/hc/log"
	hca "github.com/nickw444/homekit/bridge/accessories"
	"github.com/nickw444/homekit/bridge/mqtt"
	"github.com/nickw444/homekit/bridge/services"
	"github.com/nickw444/homekit/bridge/services/rf"
	"gopkg.in/alecthomas/kingpin.v2"
)

var log = logrus.New()

func main() {
	var (
		app        = kingpin.New("MQTTBridge", "Homekit MQTT Bridge")
		configFile = app.Flag("config", "Provide a configuration file.").Short('c').Default("bridge.conf.yml").String()

		accessCode = app.Arg("access-code", "Homekit Access code to use").Required().String()
		port       = app.Flag("port", "Port for Homekit to listen on.").Short('p').String()
		debug      = app.Flag("debug", "Enable debug output").Bool()
		hcDebug    = app.Flag("hc-debug", "Enable debug output for hc library").Bool()

		dummyMqtt         = app.Flag("dummy-mqtt", "Use a dummy MQTT instance").Bool()
		mqttBroker        = app.Flag("mqtt-broker", "MQTT Broker URL").Default("tls://127.0.0.1:8883").String()
		mqttUser          = app.Flag("mqtt-user", "MQTT Broker User").String()
		mqttPassword      = app.Flag("mqtt-password", "MQTT Password").String()
		mqttTlsSkipVerify = app.Flag("mqtt-tls-skip-verify", "Skip TLS cert verification").Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		log.Level = logrus.DebugLevel
	}

	if *hcDebug {
		hcLog.Debug.Enable()
		hcLog.Info.Enable()
	}

	var mqttClient mqtt.Client
	if *dummyMqtt {
		mqttClient = &mqtt.LoggingClient{
			Log: log.WithField("component", "logging-mqtt-client"),
		}
	} else {
		logger := log.WithField("component", "paho-mqtt-client")
		client := mqtt.NewPahoClient(*mqttBroker, *mqttUser, *mqttPassword, *mqttTlsSkipVerify, logger)
		if err := client.Connect(); err != nil {
			log.Panic(err)
		}
		mqttClient = client
	}

	config := parseConfig(*configFile)
	serviceRegistry := services.NewRegistry()
	registerServices(serviceRegistry, config.Services, mqttClient)

	// Create the bridge.
	bridge := accessory.New(accessory.Info{
		Name:         config.Name,
		Manufacturer: config.Manufacturer,
		Model:        config.Model,
	}, accessory.TypeBridge)

	// Make the accessories
	var accessories []*accessory.Accessory
	for _, accConf := range config.Accessories {
		acc := makeAccessory(mqttClient, serviceRegistry, accConf)
		accessories = append(accessories, acc.GetHCAccessory())
	}

	t, err := hc.NewIPTransport(hc.Config{Pin: *accessCode, Port: *port}, bridge, accessories...)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	log.Printf("Started server with access code: '%s' on port %s", *accessCode, *port)
	t.Start()
}

func registerServices(registry *services.Registry, services []*serviceConfig, mqttClient mqtt.Client) {
	// Register Services.
	for _, serviceConfig := range services {
		var service interface{}
		logger := log.WithField("service", serviceConfig.ID).
			WithField("id", serviceConfig.ID)

		if serviceConfig.Type == "rf" {
			rfConfig := rf.NewConfig(serviceConfig.Conf)
			service = rf.New(rfConfig, mqttClient, logger)
		} else {
			log.Panicf("Not a valid service: '%s'", serviceConfig.Type)
		}

		registry.Register(serviceConfig.ID, service)
		logger.Infof("Registered Service")
	}
}

func makeAccessory(mqttClient mqtt.Client, serviceRegistry *services.Registry,
	conf *accessoryConfig) (acc hca.HCAccessory) {
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
	} else if conf.Model == "raex-blind" {
		blindConfig := hca.NewRaexBlindConfig(conf.Conf)
		if err := blindConfig.ResolveServices(serviceRegistry); err != nil {
			log.Panic(err)
		}
		acc = hca.NewRaexBlind(conf.Serial, conf.Name, blindConfig, logger)
	} else {
		log.Panicf("Not a valid accessory model: '%s'", conf.Model)
	}

	acc.Start()

	return acc
}
