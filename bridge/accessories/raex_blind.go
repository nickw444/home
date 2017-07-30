package accessories

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"

	"github.com/nickw444/homekit/bridge/rf_service"
	svc_reg "github.com/nickw444/homekit/bridge/service_registry"
)

type RaexBlind struct {
	accessory    *accessory.Accessory
	service      *service.WindowCovering
	holdPosition *characteristic.HoldPosition
	pairingOn    *characteristic.On

	config     RaexBlindConfig
	log        *logrus.Entry
	lastAction time.Time
}

type RaexBlindConfig struct {
	remote  int
	channel int

	rfServiceId string
	rfService   *rf_service.RFService
}

const (
	raexBlindActionUp   = 254
	raexBlindActionStop = 253
	raexBlindActionDown = 252
	raexBlindActionPair = 127
)

func NewRaexBlindConfig(c map[string]interface{}) RaexBlindConfig {
	conf := RaexBlindConfig{}

	if val, ok := c["remote"]; ok {
		conf.remote = val.(int)
	} else {
		panic(fmt.Errorf("Missing configuration key: remote"))
	}

	if val, ok := c["channel"]; ok {
		conf.channel = val.(int)
	} else {
		panic(fmt.Errorf("Missing configuration key: channel"))
	}

	if val, ok := c["rf-service"]; ok {
		conf.rfServiceId = val.(string)
	} else {
		panic(fmt.Errorf("Missing configuration key: rf-service"))
	}

	return conf
}

func (r *RaexBlindConfig) ResolveServices(registry *svc_reg.ServiceRegistry) error {
	svc, err := registry.Get(r.rfServiceId)
	if err != nil {
		return err
	}

	var ok bool
	if r.rfService, ok = svc.(*rf_service.RFService); !ok {
		return fmt.Errorf("Type assertion failed.")
	}

	return nil
}

func NewRaexBlind(serial string, name string, config RaexBlindConfig, log *logrus.Entry) *RaexBlind {
	acc := accessory.New(accessory.Info{
		SerialNumber: serial,
		Name:         name,
		Model:        "raex_blind",
	}, accessory.TypeWindowCovering)

	blind := &RaexBlind{
		accessory: acc,
		log:       log,
		config:    config,
	}

	blind.service = service.NewWindowCovering()
	blind.service.TargetPosition.OnValueRemoteUpdate(blind.onTargetPositionChange)

	blind.holdPosition = characteristic.NewHoldPosition()
	blind.holdPosition.OnValueRemoteUpdate(blind.onHoldPositionChange)

	blind.pairingOn = characteristic.NewOn()
	blind.pairingOn.Description = "Pairing"
	blind.pairingOn.OnValueRemoteUpdate(blind.onPairingToggleChange)

	blind.service.AddCharacteristic(blind.holdPosition.Characteristic)
	blind.service.AddCharacteristic(blind.pairingOn.Characteristic)

	acc.AddService(blind.service.Service)

	return blind
}

func (r *RaexBlind) onHoldPositionChange(pos bool) {
	if pos {
		r.log.Debugf("Reques to hold current position")
		r.sendAction(raexBlindActionStop)
		time.AfterFunc(time.Millisecond*500, func() {
			r.service.CurrentPosition.SetValue(50)
			r.service.TargetPosition.SetValue(50)
			r.service.PositionState.SetValue(characteristic.PositionStateStopped)
		})
	}
}

func (r *RaexBlind) onPairingToggleChange(on bool) {
	if on {
		r.sendAction(raexBlindActionPair)
		time.AfterFunc(time.Second*1, func() {
			r.pairingOn.SetValue(false)
		})
	}
}

func (r *RaexBlind) onTargetPositionChange(targetPosition int) {
	var action int
	var newPosition int

	r.log.Debugf("Target position changed to: %d", targetPosition)

	if targetPosition > 70 {
		action = raexBlindActionUp
		newPosition = 100
	} else if targetPosition < 30 {
		action = raexBlindActionDown
		newPosition = 0
	} else {
		action = raexBlindActionStop
		newPosition = 50
	}

	r.sendAction(action)
	time.AfterFunc(time.Millisecond*500, func() {
		r.service.CurrentPosition.SetValue(newPosition)
		r.service.TargetPosition.SetValue(newPosition)
		r.service.PositionState.SetValue(characteristic.PositionStateStopped)
	})
}

func (r *RaexBlind) makePayload(action int) string {
	return fmt.Sprintf("%d:%d:%d:", r.config.channel, r.config.remote, action)
}

func (r *RaexBlind) sendAction(action int) {
	send := func() { r.config.rfService.Transmit("raex", r.makePayload(action)) }

	send()
	if r.lastAction.IsZero() || r.lastAction.Before(time.Now().Add(-5*time.Minute)) {
		r.log.Debugf("Position last changed at: %s. Exceeded threshold. Sending twice.", r.lastAction)
		time.AfterFunc(500*time.Millisecond, send)
	}

	r.lastAction = time.Now()
}

func (g *RaexBlind) GetHCAccessory() *accessory.Accessory {
	return g.accessory
}
