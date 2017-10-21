package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type RaexBlind struct {
	accessory    *accessory.Accessory
	service      *service.WindowCovering
	holdPosition *characteristic.HoldPosition
	pairingOn    *characteristic.On

	log              *logrus.Entry
	raexClient       *RaexBlindClient
	cancelUpdateChan chan struct{}
}

func NewRaexBlind(client *RaexBlindClient, log *logrus.Entry,
	serial string, name string) *RaexBlind {

	acc := accessory.New(accessory.Info{
		SerialNumber: serial,
		Name:         name,
		Model:        "raex-blind-2",
	}, accessory.TypeWindowCovering)

	blind := &RaexBlind{
		accessory:  acc,
		raexClient: client,
		log:        log,
	}

	blind.service = service.NewWindowCovering()

	blind.holdPosition = characteristic.NewHoldPosition()

	blind.pairingOn = characteristic.NewOn()
	blind.pairingOn.Description = "Pairing"

	blind.service.AddCharacteristic(blind.holdPosition.Characteristic)
	blind.service.AddCharacteristic(blind.pairingOn.Characteristic)
	acc.AddService(blind.service.Service)

	blind.service.TargetPosition.OnValueRemoteUpdate(blind.onTargetPositionChange)
	blind.holdPosition.OnValueRemoteUpdate(blind.onHoldPositionChange)
	blind.pairingOn.OnValueRemoteUpdate(blind.onPairingToggleChange)

	return blind
}

func (r *RaexBlind) onHoldPositionChange(pos bool) {
	if pos {
		r.log.Debugf("Request to hold current position")
		r.raexClient.Hold()
		time.AfterFunc(time.Millisecond*500, func() {
			r.service.CurrentPosition.SetValue(50)
			r.service.TargetPosition.SetValue(50)
			r.service.PositionState.SetValue(characteristic.PositionStateStopped)
		})
	}
}

func (r *RaexBlind) onTargetPositionChange(targetPosition int) {
	var newPosition int
	r.log.Debugf("Target position changed to: %d", targetPosition)

	// In order to trick Siri into always allowing us to send an Up or a down
	// command, we set the new position to be a close value to the target.
	if targetPosition > 70 {
		r.raexClient.Up()
		newPosition = 98
	} else if targetPosition < 30 {
		r.raexClient.Down()
		newPosition = 2
	} else {
		r.raexClient.Hold()
		newPosition = 50
	}

	r.setCurrentPosition(newPosition)
}

func (r *RaexBlind) onPairingToggleChange(on bool) {
	if on {
		r.raexClient.Pair()
		go time.AfterFunc(time.Second*1, func() {
			r.pairingOn.SetValue(false)
		})
	}
}

func (r *RaexBlind) setCurrentPosition(newPosition int) {
	setPosition := func(position int) {
		select {
		case <-r.cancelUpdateChan:
			return
		case <-time.After(time.Millisecond * 500):
			r.service.CurrentPosition.SetValue(position)
			r.service.TargetPosition.SetValue(position)
			r.service.PositionState.SetValue(characteristic.PositionStateStopped)
		}
	}

	if r.cancelUpdateChan != nil {
		close(r.cancelUpdateChan)
	}
	r.cancelUpdateChan = make(chan struct{})
	go setPosition(newPosition)
}
