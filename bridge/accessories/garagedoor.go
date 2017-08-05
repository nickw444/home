package accessories

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/nickw444/homekit/bridge/mqtt"
)

type GarageDoor struct {
	domain    *mqtt.Domain
	accessory *accessory.Accessory
	doorSvc   *service.GarageDoorOpener
	log       *logrus.Entry
}

func NewGarageDoor(client mqtt.Client, identifier string, name string, log *logrus.Entry) *GarageDoor {
	acc := accessory.New(accessory.Info{
		SerialNumber: identifier,
		Name:         name,
		Model:        "mqtt-garagedoor",
	}, accessory.TypeGarageDoorOpener)

	doorSvc := service.NewGarageDoorOpener()
	acc.AddService(doorSvc.Service)

	topicSvc := mqtt.NewPrefixedIDTopicService("esp", identifier)

	door := &GarageDoor{
		domain:    mqtt.NewDomain(client, topicSvc),
		accessory: acc,
		doorSvc:   doorSvc,
		log:       log,
	}

	return door
}

func (g *GarageDoor) Start() {
	g.doorSvc.TargetDoorState.OnValueRemoteUpdate(g.onTargetDoorStateChange)

	// Subscribe to the door state changing
	g.domain.Subscribe("status", g.handleDoorStatusChange)

	// Get the current state
	g.domain.Republish()
}

func (g *GarageDoor) onTargetDoorStateChange(target int) {
	if g.doorSvc.CurrentDoorState.GetValue() == characteristic.CurrentDoorStateClosed {
		g.doorSvc.CurrentDoorState.SetValue(characteristic.CurrentDoorStateOpening)
	} else {
		g.doorSvc.CurrentDoorState.SetValue(characteristic.CurrentDoorStateClosing)
	}

	g.domain.Publish("trigger", "")
}

func (g *GarageDoor) handleDoorStatusChange(msg string) {
	g.log.Infof("Door Status Changed to %s\n", msg)
	status, err := newDoorStatus(msg)
	if err != nil {
		g.log.Error(err)
		return
	}

	currentState := characteristic.CurrentDoorStateOpen
	targetState := characteristic.TargetDoorStateOpen

	if status == doorOpen {
		currentState = characteristic.CurrentDoorStateOpen
		targetState = characteristic.TargetDoorStateOpen
	} else if status == doorClosed {
		currentState = characteristic.CurrentDoorStateClosed
		targetState = characteristic.TargetDoorStateClosed
	}

	g.doorSvc.CurrentDoorState.SetValue(currentState)
	g.doorSvc.TargetDoorState.SetValue(targetState)
}

// GetHCAccessory returns the homecontrol accessory.
func (g *GarageDoor) GetHCAccessory() *accessory.Accessory {
	return g.accessory
}

// doorStatus represents a concrete type for door status
type doorStatus int

const (
	doorOpen doorStatus = iota
	doorClosed
	doorUnknown
)

func newDoorStatus(val string) (status doorStatus, err error) {
	if val == "OPEN" {
		status = doorOpen
	} else if val == "CLOSED" {
		status = doorClosed
	} else if val == "UNKNOWN" {
		status = doorUnknown
	} else {
		err = fmt.Errorf("Unknown door status %s", val)
	}
	return
}
