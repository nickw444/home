package device

import (
	"github.com/nickw444/miio-go/capability"
	"github.com/nickw444/miio-go/common"
)

type PowerPlug struct {
	Device
	*capability.Power
}

func NewPowerPlug(device Device) *PowerPlug {
	dev := &PowerPlug{
		Device: device,
		Power:  capability.NewPower(device, device.Outbound()),
	}
	go dev.refresh()
	return dev
}

func (p *PowerPlug) refresh() {
	for range p.RefreshThrottle() {
		_ = p.Power.Update()
	}

	common.Log.Debug("Device refresh closed.")
}
