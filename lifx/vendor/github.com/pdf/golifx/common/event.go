package common

// EventNewDevice is emitted by a Client or Group when it discovers a new Device
type EventNewDevice struct {
	Device Device
}

// EventExpiredDevice is emitted by a Client or Group when a Device is no longer
// known
type EventExpiredDevice struct {
	Device Device
}

// EventUpdateLabel is emitted by a Device or Group when its label is updated
type EventUpdateLabel struct {
	Label string
}

// EventUpdatePower is emitted by a Device or Group when its power state is
// updated
type EventUpdatePower struct {
	Power bool
}

// EventUpdateColor is emitted by a Light or Group when its Color is updated
type EventUpdateColor struct {
	Color Color
}

// EventNewLocation is emitted by a Client when it discovers a new Location
type EventNewLocation struct {
	Location Location
}

// EventExpiredLocation is emitted by a Client or Group when a Location is no longer
// known
type EventExpiredLocation struct {
	Location Location
}

// EventNewGroup is emitted by a Client when it discovers a new Group
type EventNewGroup struct {
	Group Group
}

// EventExpiredGroup is emitted by a Client or Group when a Group is no longer
// known
type EventExpiredGroup struct {
	Group Group
}
