package common

import "time"

// Light represents a LIFX light device
type Light interface {
	// SetColor changes the color of the light, transitioning over the specified
	// duration
	SetColor(color Color, duration time.Duration) error
	// GetColor requests the current color of the light
	GetColor() (Color, error)
	// CachedColor returns the last known color of the light
	CachedColor() Color
	// SetPowerDuration sets the power of the light, transitioning over the
	// speficied duration, state is true for on, false for off.
	SetPowerDuration(state bool, duration time.Duration) error

	// Light is a superset of the Device interface
	Device
}
