package common

type PowerState string

const (
	PowerStateUnknown PowerState = ""
	PowerStateOn                 = "on"
	PowerStateOff                = "off"
)
