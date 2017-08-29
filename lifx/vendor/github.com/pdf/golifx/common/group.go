package common

import "time"

// Group represents a group of LIFX devices
type Group interface {
	// ID returns a base64 encoding of the device ID
	ID() string

	// Label returns the label for the group
	GetLabel() string

	// Devices returns the devices in the group
	Devices() []Device

	// Lights returns the lights in the group
	Lights() []Light

	// Returns the power state of the group, true if any members are on, false
	// if all members off. Returns error on communication errors.
	GetPower() (bool, error)

	// Returns the average color of lights in the group. Returns error on
	// communication error.
	GetColor() (Color, error)

	// SetColor requests a change of color for all devices in the group that
	// support color changes, transitioning over the specified duration
	SetColor(color Color, duration time.Duration) error
	// SetPower sets the power of devices in the group that support power
	// changes, state is true for on, false for off.
	SetPower(state bool) error
	// SetPowerDuration sets the power of devices in the group that support
	// power changes, transitioning over the speficied duration, state is true
	// for on, false for off.
	SetPowerDuration(state bool, duration time.Duration) error

	// Device is a SubscriptionTarget
	SubscriptionTarget
}
