package accessories

import (
	"github.com/brutella/hc/accessory"
)

type HCAccessory interface {
	// GetHCAccessory returns the associated hc/homekit
	// accessory.
	GetHCAccessory() *accessory.Accessory

	// Start should subscribe to accessories and bind any
	// callbacks required for operation
	Start()
}
