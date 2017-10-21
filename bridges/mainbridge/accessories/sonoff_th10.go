package accessories

import (
	"strconv"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
	"github.com/nickw444/homekit/bridges/lib/mqtt"
)

type SonoffTH10 struct {
	domain      *mqtt.Domain
	humiditySvc *service.HumiditySensor
	accessory   *accessory.Thermometer
}

const (
	topicEndpointTemperature = "temperature"
	topicEndpointHumidity    = "humidity"
)

// NewSonoffTH10 creates a new Thermometer device.
func NewSonoffTH10(client mqtt.Client, identifier string, name string) *SonoffTH10 {
	acc := accessory.NewTemperatureSensor(accessory.Info{
		SerialNumber: identifier,
		Name:         name,
		Model:        "sonoff-th10",
	}, 0, -40, 100, 0.1)

	humidity := service.NewHumiditySensor()
	acc.AddService(humidity.Service)

	topicSvc := mqtt.NewPrefixedIDTopicService("esp", identifier)

	sonoff := &SonoffTH10{
		domain:      mqtt.NewDomain(client, topicSvc),
		accessory:   acc,
		humiditySvc: humidity,
	}

	return sonoff
}

func (s *SonoffTH10) Start() {
	// Setup the listener
	s.domain.Subscribe(topicEndpointTemperature, s.handleTemperatureReceived)
	s.domain.Subscribe(topicEndpointHumidity, s.handleHumidityReceived)

	// Republish to get the current reading.
	s.domain.Republish()
}

// GetHCAccessory returns the homekit accessory.
func (s *SonoffTH10) GetHCAccessory() *accessory.Accessory {
	return s.accessory.Accessory
}

func (s *SonoffTH10) handleTemperatureReceived(msg string) {
	temperature, _ := strconv.ParseFloat(msg, 64)
	s.accessory.TempSensor.CurrentTemperature.SetValue(temperature)
}

func (s *SonoffTH10) handleHumidityReceived(msg string) {
	humidity, _ := strconv.ParseFloat(msg, 64)
	s.humiditySvc.CurrentRelativeHumidity.SetValue(humidity)
}
