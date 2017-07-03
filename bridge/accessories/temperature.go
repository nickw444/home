package accessories

import (
	"strconv"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
	"github.com/nickw444/homekit/bridge/mqtt"
)

type Thermometer struct {
	domain      *mqtt.Domain
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

	topicSvc := mqtt.NewPrefixedIDTopicService("esp", identifier)

	sonoff := &Thermometer{
		domain:      mqtt.NewDomain(client, topicSvc),
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

func (t *Thermometer) handleTemperatureReceived(msg string) {
	temperature, _ := strconv.ParseFloat(msg, 64)
	t.accessory.TempSensor.CurrentTemperature.SetValue(temperature)
}

func (t *Thermometer) handleHumidityReceived(msg string) {
	humidity, _ := strconv.ParseFloat(msg, 64)
	t.humiditySvc.CurrentRelativeHumidity.SetValue(humidity)
}
