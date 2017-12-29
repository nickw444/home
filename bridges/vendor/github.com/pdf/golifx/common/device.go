package common

// Device represents a generic LIFX device
type Device interface {
	// Returns the ID for the device
	ID() uint64

	// GetLabel gets the label for the device
	GetLabel() (string, error)
	// SetLabel sets the label for the device
	SetLabel(label string) error
	// GetPower requests the current power state of the device, true for on,
	// false for off
	GetPower() (bool, error)
	// CachedPower returns the last known power state of the device, true for
	// on, false for off
	CachedPower() bool
	// SetPower sets the power state of the device, true for on, false for off
	SetPower(state bool) error
	// GetFirmwareVersion returns the firmware version of the device
	GetFirmwareVersion() (string, error)
	// CachedFirmwareVersion returns the last known firmware version of the
	// device
	CachedFirmwareVersion() string
	// GetProductName returns the product name of the device
	GetProductName() (string, error)

	// Device is a SubscriptionTarget
	SubscriptionTarget
}
