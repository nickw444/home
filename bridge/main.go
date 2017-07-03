package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hcLog "github.com/brutella/hc/log"

	"github.com/Sirupsen/logrus"
	hca "github.com/nickw444/homekit/bridge/accessories"
	"github.com/nickw444/homekit/bridge/mqtt"
	"gopkg.in/alecthomas/kingpin.v2"

	"crypto/tls"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"

	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
)

var log = logrus.New()

func main() {
	var (
		app        = kingpin.New("MQTTBridge", "Homekit MQTT Bridge")
		configFile = app.Flag("config", "Provide a configuration file.").Default("bridge.conf.yml").String()

		accessCode = app.Arg("accessCode", "Homekit Access code to use").Required().String()
		port       = app.Flag("port", "Port for Homekit to listen on.").String()
		debug      = app.Flag("debug", "Enable debug output").Bool()

		dummyMqtt         = app.Flag("dummy-mqtt", "Use a dummy MQTT instance").Bool()
		mqttBroker        = app.Flag("mqtt-broker", "MQTT Broker URL").Default("tls://127.0.0.1:8883").String()
		mqttUser          = app.Flag("mqtt-user", "MQTT Broker User").String()
		mqttPassword      = app.Flag("mqtt-password", "MQTT Password").String()
		mqttTlsSkipVerify = app.Flag("mqtt-tls-skip-verify", "Skip TLS cert verification").Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		log.Level = logrus.DebugLevel
		hcLog.Debug.Enable()
		hcLog.Info.Enable()
	}

	var mqttClient mqtt.Client
	if *dummyMqtt {
		mqttClient = &mqtt.LoggingClient{
			Log: log.WithField("service", "logging-mqtt-service"),
		}
	} else {
		mqttClient = makePahoMqttClient(*mqttBroker, *mqttUser, *mqttPassword, *mqttTlsSkipVerify)
	}

	config := parseConfig(*configFile)
	// Create the bridge.
	bridge := accessory.New(accessory.Info{
		Name:         config.Name,
		Manufacturer: config.Manufacturer,
		Model:        config.Model,
	}, accessory.TypeBridge)

	// Make the accessories
	var accessories []*accessory.Accessory
	for _, accConf := range config.Accessories {
		log.Println(accConf)
		acc := makeAccessory(mqttClient, accConf)
		accessories = append(accessories, acc.GetHCAccessory())
	}

	t, err := hc.NewIPTransport(hc.Config{Pin: *accessCode, Port: *port}, bridge, accessories...)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	log.Printf("Started server with access code: %s", *accessCode)
	t.Start()
}

func makeAccessory(mqttClient mqtt.Client, conf *accessoryConfig) hca.HCAccessory {
	logger := log.WithField("accessory", conf.Model)

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
		return hca.NewBlind(conf.Serial, conf.Name, logger)
	}

	log.Panicf("Not a valid accessory model: '%s'", conf.Model)
	return nil
}

func makePahoMqttClient(broker string, user string, password string, tlsSkipVerify bool) mqtt.Client {
	clientOptions := pahoMqtt.NewClientOptions().AddBroker(broker)
	if user != "" {
		clientOptions = clientOptions.SetUsername(user)
	}
	if password != "" {
		clientOptions = clientOptions.SetPassword(password)
	}
	if tlsSkipVerify {
		clientOptions = clientOptions.SetTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}

	pahoClient := pahoMqtt.NewClient(clientOptions)
	if token := pahoClient.Connect(); token.Wait() && token.Error() != nil {
		log.Panic(token.Error())
	}

	return &mqtt.PahoClient{
		Client: pahoClient,
	}
}

type accessoryConfig struct {
	Model  string
	Serial string
	Name   string
	Conf   map[string]interface{}
}

type bridgeConfig struct {
	Name         string
	Manufacturer string
	Model        string
	Accessories  []*accessoryConfig
}

func parseConfig(filename string) *bridgeConfig {
	file, err := os.Open(filename) // For read access.
	if err != nil {
		log.Panic(err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}

	var conf bridgeConfig
	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		log.Panic(err)
	}

	return &conf
}
