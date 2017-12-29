package device

import (
	"time"

	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol/v2/packet"
)

type GenericDevice interface {
	common.Device
	Handle(*packet.Packet)
	Close() error
	Seen() time.Time
	SetSeen(time.Time)
	Provisional() bool
	SetProvisional(bool)
	SetStatePower(*packet.Packet) error
	SetStateLabel(*packet.Packet) error
	SetStateLocation(*packet.Packet) error
	SetStateGroup(*packet.Packet) error
	GetLocation() (string, error)
	CachedLocation() string
	GetGroup() (string, error)
	CachedGroup() string
	GetProduct() (*Product, error)
	ResetLimiter()
}
