package accessories

import (
	"github.com/brutella/hc/accessory"
	"github.com/nickw444/homekit/bridge/mqtt"
)

type SonoffSwitchConfig struct {
	fwVersion int
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
	conf.fwVersion = 2

	if val, ok := c["fw-version"]; ok {
		conf.fwVersion = val.(int)
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
	if switchConfig.fwVersion == 1 {
		prefix = "device"
	} else {
		prefix = "esp"
	}

	topicSvc := mqtt.NewPrefixedIDTopicService(prefix, identifier)
	domain := mqtt.NewDomain(client, topicSvc)

	sonoff := &SonoffSwitch{
		domain:          domain,
		switchAccessory: acc,
	}

	return sonoff
}

func (s *SonoffSwitch) Start() {
	s.switchAccessory.Switch.On.OnValueRemoteUpdate(s.onSwitchTargetStateChange)

	// Setup the listener
	s.domain.Subscribe(topicEndpointRelayState, s.handleRelayStateMsg)

	// Republish it's existing status so that we can update the switch.
	s.domain.Republish()

}

func (s *SonoffSwitch) onSwitchTargetStateChange(b bool) {
	if b {
		s.setState(SonoffRelayStateOn)
	} else {
		s.setState(SonoffRelayStateOff)
	}
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
