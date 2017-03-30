package accessories

import (
	"github.com/nickw444/homekit/bridge/mqtt_domain"
	"github.com/nickw444/homekit/bridge/topic_service"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type GarageDoor struct {
	domain    *mqtt_domain.MQTTDomain
	accessory *accessory.Accessory
	doorSvc   *service.GarageDoorOpener
}

const (
	topicEndpointTrigger   = "trigger"
	topicEndpointDoorState = "state"
)

func NewGarageDoor(client mqtt.Client, identifier string, name string) *GarageDoor {
	acc := accessory.New(accessory.Info{
		SerialNumber: identifier,
		Name:         name,
		Model:        "mqtt-garagedoor",
	}, accessory.TypeGarageDoorOpener)

	doorSvc := service.NewGarageDoorOpener()
	acc.AddService(doorSvc.Service)

	topicSvc := topic_service.NewPrefixedIDTopicService("esp", identifier)
	door := &GarageDoor{
		domain:    mqtt_domain.NewMQTTDomain(client, topicSvc),
		accessory: acc,
		doorSvc:   doorSvc,
	}

	doorSvc.TargetDoorState.OnValueRemoteUpdate(func(target int) {
		if doorSvc.CurrentDoorState.GetValue() == characteristic.CurrentDoorStateClosed {
			doorSvc.CurrentDoorState.SetValue(characteristic.CurrentDoorStateOpening)
		} else {
			doorSvc.CurrentDoorState.SetValue(characteristic.CurrentDoorStateClosing)
		}

		door.domain.Publish(topicEndpointTrigger, "")
	})

	// Subscribe to the door state changing
	door.domain.Subscribe(topicEndpointDoorState, door.handleRelayStateMsg)

	// Get the current state
	door.domain.Republish()

	return door
}

func (g *GarageDoor) handleRelayStateMsg(c mqtt.Client, msg mqtt.Message) {
	m := string(msg.Payload())

	currentState := characteristic.CurrentDoorStateOpen
	targetState := characteristic.TargetDoorStateOpen

	if m == "1" {
		currentState = characteristic.CurrentDoorStateOpen
		targetState = characteristic.TargetDoorStateOpen
	} else if m == "0" {
		currentState = characteristic.CurrentDoorStateClosed
		targetState = characteristic.TargetDoorStateClosed
	}

	g.doorSvc.CurrentDoorState.SetValue(currentState)
	g.doorSvc.TargetDoorState.SetValue(targetState)
}

func (g *GarageDoor) GetHCAccessory() *accessory.Accessory {
	return g.accessory
}
