package device

import (
	"time"

	"github.com/nickw444/miio-go/common"
	"github.com/nickw444/miio-go/device/product"
	"github.com/nickw444/miio-go/protocol/packet"
	"github.com/nickw444/miio-go/protocol/transport"
)

type Device interface {
	common.Device

	Handle(*packet.Packet) error
	Close() error
	Seen() time.Time
	Provisional() bool
	SetProvisional(bool)
	GetProduct() (product.Product, error)
	Discover() error
	RefreshThrottle() <-chan struct{}
	Outbound() transport.Outbound
}
