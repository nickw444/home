package common

type EventNewDevice struct {
	Device Device
}

type EventExpiredDevice struct {
	Device Device
}

type EventUpdatePower struct {
	PowerState PowerState
}
