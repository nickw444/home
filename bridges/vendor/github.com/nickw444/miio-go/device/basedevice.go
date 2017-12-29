package device

import (
	"time"

	"github.com/nickw444/miio-go/common"
	"github.com/nickw444/miio-go/device/product"
	"github.com/nickw444/miio-go/protocol/packet"
	"github.com/nickw444/miio-go/subscription"
	"github.com/nickw444/miio-go/protocol/transport"
)

// baseDevice implements the Device interface.
type baseDevice struct {
	subscription.SubscriptionTarget

	refreshThrottle RefreshThrottle
	outbound        transport.Outbound

	product     product.Product
	id          uint32
	provisional bool
	seen        time.Time
}

type Info struct {
	FirmwareVersion string `json:"fw_ver"`
	HardwareVersion string `json:"hw_ver"`
	MacAddress      string `json:"mac"`
	Model           string `json:"model"`
}

type InfoResponse struct {
	Result Info  `json:"result"`
	ID     int32 `json:"ID"`
}

func New(deviceId uint32, transport transport.Outbound, seen time.Time) Device {
	throttle := NewRefreshThrottle(time.Second * 5)
	b := &baseDevice{
		SubscriptionTarget: subscription.NewTarget(),

		refreshThrottle: throttle,
		outbound:        transport,
		id:              deviceId,
		seen:            seen,
	}
	b.init()
	return b
}

func (b *baseDevice) init() {
	b.product = product.Unknown
	b.provisional = true
}

func (b *baseDevice) ID() uint32 {
	return b.id
}

func (b *baseDevice) GetLabel() (string, error) {
	return "", nil
}

func (b *baseDevice) Handle(pkt *packet.Packet) error {
	common.Log.Infof("Handling packet at base_device")
	b.seen = pkt.Meta.DecodeTime
	return b.outbound.Handle(pkt)
}

func (b *baseDevice) Close() error {
	err := b.SubscriptionTarget.CloseAllSubscriptions()
	b.refreshThrottle.Close()
	return err
}

func (b *baseDevice) Seen() time.Time {
	return b.seen
}

func (b *baseDevice) Provisional() bool {
	return b.provisional
}

func (b *baseDevice) SetProvisional(provisional bool) {
	b.provisional = provisional
}

func (b *baseDevice) GetProduct() (product.Product, error) {
	resp := InfoResponse{}
	err := b.outbound.CallAndDeserialize("miIO.info", nil, &resp)
	if err != nil {
		return product.Unknown, err
	}

	return product.GetModel(resp.Result.Model)
}

func (b *baseDevice) Discover() error {
	return b.outbound.Send(packet.NewHello())
}

func (b *baseDevice) NewSubscription() (subscription.Subscription, error) {
	sub, err := b.SubscriptionTarget.NewSubscription()
	b.refreshThrottle.Start()
	return sub, err
}

func (b *baseDevice) RemoveSubscription(s subscription.Subscription) (err error) {
	err = b.SubscriptionTarget.RemoveSubscription(s)
	if !b.HasSubscribers() {
		b.refreshThrottle.Stop()
	}
	return
}

func (b *baseDevice) RefreshThrottle() <-chan struct{} {
	return b.refreshThrottle.Chan()
}

func (b *baseDevice) Outbound() transport.Outbound {
	return b.outbound
}
