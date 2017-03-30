package accessories

import (
	"strconv"

	"github.com/nickw444/homekit/bridge/mqtt_domain"
	"github.com/nickw444/homekit/bridge/topic_service"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Thermometer struct {
	domain      *mqtt_domain.MQTTDomain
	humiditySvc *service.HumiditySensor
	accessory   *accessory.Thermometer
}

const (
	topicEndpointTemperature = "temperature"
	topicEndpointHumidity    = "humidity"
)

// NewThermometer creates a new Thermometer device.
func NewThermometer(client mqtt.Client, identifier string, name string) *Thermometer {

	acc := accessory.NewTemperatureSensor(accessory.Info{
		SerialNumber: identifier,
		Name:         name,
		Model:        "sonoff-th10",
	}, 0, -40, 100, 0.1)

	humidity := service.NewHumiditySensor()
	acc.AddService(humidity.Service)

	topicSvc := topic_service.NewPrefixedIDTopicService("esp", identifier)

	sonoff := &Thermometer{
		domain:      mqtt_domain.NewMQTTDomain(client, topicSvc),
		accessory:   acc,
		humiditySvc: humidity,
	}

	// Setup the listener
	sonoff.domain.Subscribe(topicEndpointTemperature, sonoff.handleTemperatureReceived)
	sonoff.domain.Subscribe(topicEndpointHumidity, sonoff.handleHumidityReceived)

	// Republish to get the current reading.
	sonoff.domain.Republish()

	return sonoff
}

// GetHCAccessory returns the homekit accessory.
func (t *Thermometer) GetHCAccessory() *accessory.Accessory {
	return t.accessory.Accessory
}

func (t *Thermometer) handleTemperatureReceived(c mqtt.Client, msg mqtt.Message) {
	m := string(msg.Payload())
	temperature, _ := strconv.ParseFloat(m, 64)
	t.accessory.TempSensor.CurrentTemperature.SetValue(temperature)
}

func (t *Thermometer) handleHumidityReceived(c mqtt.Client, msg mqtt.Message) {
	m := string(msg.Payload())
	humidity, _ := strconv.ParseFloat(m, 64)
	t.humiditySvc.CurrentRelativeHumidity.SetValue(humidity)
}
