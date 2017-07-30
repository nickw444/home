package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hcLog "github.com/brutella/hc/log"

	"github.com/Sirupsen/logrus"
	hca "github.com/nickw444/homekit/bridge/accessories"
	"github.com/nickw444/homekit/bridge/mqtt"
	rf_service "github.com/nickw444/homekit/bridge/rf_service"
	svc_reg "github.com/nickw444/homekit/bridge/service_registry"
	"gopkg.in/alecthomas/kingpin.v2"

	"os"
)

var log = logrus.New()

func main() {
	var (
		app        = kingpin.New("MQTTBridge", "Homekit MQTT Bridge")
		configFile = app.Flag("config", "Provide a configuration file.").Default("bridge.conf.yml").String()

		accessCode = app.Arg("accessCode", "Homekit Access code to use").Required().String()
		port       = app.Flag("port", "Port for Homekit to listen on.").String()
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
	serviceRegistry := svc_reg.NewServiceRegistry()
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

func registerServices(registry *svc_reg.ServiceRegistry, services []*serviceConfig, mqttClient mqtt.Client) {
	// Register Services.
	for _, serviceConfig := range services {
		var service interface{}
		logger := log.WithField("service", serviceConfig.ID).
			WithField("id", serviceConfig.ID)

		if serviceConfig.Type == "rf" {
			rfConfig := rf_service.NewConfig(serviceConfig.Conf)
			service = rf_service.New(rfConfig, mqttClient, logger)
		} else {
			log.Panicf("Not a valid service: '%s'", serviceConfig.Type)
		}

		registry.Register(serviceConfig.ID, service)
		logger.Infof("Registered Service")
	}
}

func makeAccessory(mqttClient mqtt.Client, serviceRegistry *svc_reg.ServiceRegistry,
	conf *accessoryConfig) hca.HCAccessory {

	logger := log.WithField("accessory", conf.Model).
		WithField("serial", conf.Serial)

	logger.Infof("Loading accessory...")

	if conf.Model == "sonoff-switch" {
		switchConfig := hca.NewSonoffSwitchConfig(conf.Conf)
		return hca.NewSonoffSwitch(switchConfig, mqttClient, conf.Serial, conf.Name)
	} else if conf.Model == "garagedoor" {
		return hca.NewGarageDoor(mqttClient, conf.Serial, conf.Name, logger)
	} else if conf.Model == "sonoff-th10" {
		return hca.NewThermometer(mqttClient, conf.Serial, conf.Name)
	} else if conf.Model == "latch-lock" {
		lockConfig := hca.NewLatchLockConfig(conf.Conf)
		return hca.NewLatchLock(lockConfig, mqttClient, conf.Serial, conf.Name, logger)
	} else if conf.Model == "raex-blind" {
		blindConfig := hca.NewRaexBlindConfig(conf.Conf)
		if err := blindConfig.ResolveServices(serviceRegistry); err != nil {
			log.Panic(err)
		}
		return hca.NewRaexBlind(conf.Serial, conf.Name, blindConfig, logger)
	}

	log.Panicf("Not a valid accessory model: '%s'", conf.Model)
	return nil
}
