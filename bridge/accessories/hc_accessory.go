package accessories

import (
	"github.com/brutella/hc/accessory"
)

type HCAccessory interface {
	GetHCAccessory() *accessory.Accessory
}
