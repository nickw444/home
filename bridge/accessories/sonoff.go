package accessories

import (
	"github.com/brutella/hc/accessory"
	"github.com/nickw444/homekit/bridge/mqtt"
)

type SonoffSwitchConfig struct {
	FWVersion int
}

type SonoffSwitch struct {
	domain          *mqtt.Domain
	switchAccessory *accessory.Switch
}

type SonoffRelayState int

const (
	SonoffRelayStateOn = iota
	SonoffRelayStateOff
)

const (
	topicEndpointRelaySet   = "relay/set"
	topicEndpointRelayState = "relay"
)

func NewSonoffSwitchConfig(c map[string]interface{}) SonoffSwitchConfig {
	conf := SonoffSwitchConfig{}
	conf.FWVersion = 2

	if val, ok := c["fw-version"]; ok {
		conf.FWVersion = val.(int)
	}

	return conf
}

func NewSonoffSwitch(switchConfig SonoffSwitchConfig, client mqtt.Client,
	identifier string, name string) *SonoffSwitch {

	acc := accessory.NewSwitch(accessory.Info{
		SerialNumber: identifier,
		Name:         name,
		Model:        "sonoff-switch",
	})

	var prefix string
	if switchConfig.FWVersion == 1 {
		prefix = "device"
	} else {
		prefix = "esp"
	}

	topicSvc := mqtt.NewPrefixedIDTopicService(prefix, identifier)

	sonoff := &SonoffSwitch{
		domain:          mqtt.NewDomain(client, topicSvc),
		switchAccessory: acc,
	}

	acc.Switch.On.OnValueRemoteUpdate(func(b bool) {
		if b {
			sonoff.setState(SonoffRelayStateOn)
		} else {
			sonoff.setState(SonoffRelayStateOff)
		}
	})

	// Setup the listener
	sonoff.domain.Subscribe(topicEndpointRelayState, sonoff.handleRelayStateMsg)

	// Republish it's existing status so that we can update the switch.
	sonoff.domain.Republish()

	return sonoff

}

func (s *SonoffSwitch) handleRelayStateMsg(msg string) {
	if msg == "1" {
		s.switchAccessory.Switch.On.SetValue(true)
	} else if msg == "0" {
		s.switchAccessory.Switch.On.SetValue(false)
	}
}

func (s *SonoffSwitch) setState(state SonoffRelayState) {
	msg := ""

	if state == SonoffRelayStateOff {
		msg = "0"
	} else if state == SonoffRelayStateOn {
		msg = "1"
	}

	s.domain.Publish(topicEndpointRelaySet, msg)
}

func (s *SonoffSwitch) GetHCAccessory() *accessory.Accessory {
	return s.switchAccessory.Accessory
}
