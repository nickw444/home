package accessories

import (
	"github.com/Sirupsen/logrus"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type Blind struct {
	accessory *accessory.Accessory
	service   *service.WindowCovering
}

func NewBlind(serial string, name string, log *logrus.Entry) *Blind {
	acc := accessory.New(accessory.Info{
		SerialNumber: serial,
		Name:         name,
		Model:        "raex-blind",
	}, accessory.TypeWindowCovering)

	svc := service.NewWindowCovering()
	holdPosition := characteristic.NewHoldPosition()
	svc.AddCharacteristic(holdPosition.Characteristic)

	acc.AddService(svc.Service)

	blind := &Blind{
		accessory: acc,
	}

	return blind
}

func (g *Blind) GetHCAccessory() *accessory.Accessory {
	return g.accessory
}
