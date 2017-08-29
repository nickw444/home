package common

import (
	"time"
)

// Protocol defines the interface between the Client and a protocol
// implementation
type Protocol interface {
	SubscriptionTarget

	// GetLocations returns a slice of all locations known to the protocol, or
	// ErrNotFound if no locations are currently known.
	GetLocations() (locations []Location, err error)
	// GetLocation looks up a location by its `id`
	GetLocation(id string) (Location, error)
	// GetGroups returns a slice of all groups known to the protocol, or
	// ErrNotFound if no locations are currently known.
	GetGroups() (locations []Group, err error)
	// GetGroup looks up a group by its `id`
	GetGroup(id string) (Group, error)
	// GetDevices returns a slice of all devices known to the protocol, or
	// ErrNotFound if no devices are currently known.
	GetDevices() (devices []Device, err error)
	// GetDevice looks up a device by its `id`
	GetDevice(id uint64) (Device, error)
	// Discover initiates device discovery, this may be a noop in some future
	// protocol versions.  This is called immediately when the client connects
	// to the protocol
	Discover() error
	// SetTimeout attaches the client timeout to the protocol
	SetTimeout(timeout *time.Duration)
	// SetRetryInterval attaches the client retry interval to the protocol
	SetRetryInterval(retryInterval *time.Duration)
	// Close closes the protocol driver, no further communication with the
	// protocol is possible
	Close() error
	// NewSubscription returns a *Subscription for a Client to obtain
	// events from the Protocol

	// SetPower sets the power state globally, on all devices
	SetPower(state bool) error
	// SetPowerDuration sets the power state globally, on all lights, over the
	// specified duration
	SetPowerDuration(state bool, duration time.Duration) error
	// SetColor changes the color globally, on all lights, over the specified
	// duration
	SetColor(color Color, duration time.Duration) error
}
