package device

import (
	"fmt"

	"github.com/nickw444/miio-go/device/product"
)

// Classify determines the underlying product of the device and returns an
// appropriate device implementation.
func Classify(dev Device) (Device, error) {
	if !dev.Provisional() {
		return dev, nil
	}

	p, err := dev.GetProduct()
	if err != nil {
		return nil, err
	}

	defer dev.SetProvisional(false)

	switch p {
	case product.Yeelight:
		return NewYeelight(dev), nil
	case product.PowerPlug:
		return NewPowerPlug(dev), nil
	default:
		return nil, fmt.Errorf("Classify: Unknown device type")
	}
}
