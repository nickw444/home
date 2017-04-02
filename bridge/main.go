package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"

	"github.com/Sirupsen/logrus"
	hca "github.com/nickw444/homekit/bridge/accessories"
	"gopkg.in/alecthomas/kingpin.v2"

	"crypto/tls"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var log = logrus.New()

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

func makeAccessory(mqttClient mqtt.Client, conf *accessoryConfig) hca.HCAccessory {
	if conf.Model == "sonoff-switch" {
		switchConfig := hca.NewSonoffSwitchConfig(conf.Conf)
		return hca.NewSonoffSwitch(switchConfig, mqttClient, conf.Serial, conf.Name)
	} else if conf.Model == "garagedoor" {
		return hca.NewGarageDoor(mqttClient, conf.Serial, conf.Name)
	} else if conf.Model == "sonoff-th10" {
		return hca.NewThermometer(mqttClient, conf.Serial, conf.Name)
	}

	log.Panicf("Not a valid accessory model: '%s'", conf.Model)
	return nil
}

func main() {

	var (
		app        = kingpin.New("MQTTBridge", "Homekit MQTT Bridge")
		configFile = app.Flag("config", "Provide a configuration file.").Default("bridge.conf.yml").String()

		accessCode   = app.Arg("accessCode", "Homekit Access code to use").Required().String()
		port         = app.Flag("port", "Port for Homekit to listen on.").String()
		mqttBroker   = app.Flag("mqttBroker", "MQTT Broker URL").Default("tls://127.0.0.1:8883").String()
		mqttUser     = app.Flag("mqttUser", "MQTT Broker User").String()
		mqttPassword = app.Flag("mqttPassword", "MQTT Password").String()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	clientOptions := mqtt.NewClientOptions().AddBroker(*mqttBroker).SetTLSConfig(tlsConfig)
	if *mqttUser != "" {
		clientOptions = clientOptions.SetUsername(*mqttUser)
	}

	if *mqttPassword != "" {
		clientOptions = clientOptions.SetPassword(*mqttPassword)
	}

	client := mqtt.NewClient(clientOptions)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Panic(token.Error())
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
		acc := makeAccessory(client, accConf)
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
