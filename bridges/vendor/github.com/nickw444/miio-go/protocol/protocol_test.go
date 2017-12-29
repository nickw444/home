package protocol

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/nickw444/miio-go/device"
	"github.com/nickw444/miio-go/mocks"
	"github.com/nickw444/miio-go/protocol/packet"
	"github.com/nickw444/miio-go/protocol/transport"
	smocks "github.com/nickw444/miio-go/subscription/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Protocol_SetUp() (tt struct {
	clk                *clock.Mock
	transport          *mockTransport
	deviceFactory      DeviceFactory
	cryptoFactory      CryptoFactory
	subscriptionTarget *smocks.SubscriptionTarget
	protocol           *protocol
	devices            []*mocks.Device
}) {
	tt.clk = clock.NewMock()
	tt.transport = &mockTransport{new(mocks.Inbound)}
	tt.subscriptionTarget = new(smocks.SubscriptionTarget)
	tt.deviceFactory = func(deviceId uint32, outbound transport.Outbound, seen time.Time) device.Device {
		d := &mocks.Device{}
		d.On("Discover").Return(nil)
		//d.On("Packets").Return(nil)
		tt.devices = append(tt.devices, d)
		return d
	}
	tt.cryptoFactory = func(deviceID uint32, deviceToken []byte, initialStamp uint32, stampTime time.Time) (packet.Crypto, error) {
		return nil, nil
	}
	tt.protocol = newProtocol(tt.clk, tt.transport, tt.deviceFactory, tt.cryptoFactory, tt.subscriptionTarget)
	return
}

// Ensure that the broadcast device has Discover called on it.
func TestProtocol_Discover(t *testing.T) {
	tt := Protocol_SetUp()

	err := tt.protocol.Discover()
	assert.NoError(t, err)
	tt.devices[0].AssertCalled(t, "Discover")
}

// Ensure that inbound's Packets method is called.
func TestProtocol_dispatcher(t *testing.T) {
	tt := Protocol_SetUp()
	wg := sync.WaitGroup{}
	wg.Add(1)

	ch := make(chan *packet.Packet)
	// Hack to convert the channel to a read-only channel (what the mock expects)
	ro := func(c chan *packet.Packet) <-chan *packet.Packet {
		return c
	}

	tt.transport.inbound.On("Packets").Return(ro(ch)).Run(func(args mock.Arguments) {
		wg.Done()
	})
	tt.protocol.start()
	wg.Wait()
	tt.transport.inbound.AssertExpectations(t)
}

type mockTransport struct {
	inbound *mocks.Inbound
}

func (m *mockTransport) Inbound() transport.Inbound {
	return m.inbound
}

func (*mockTransport) NewOutbound(crypto packet.Crypto, dest net.Addr) transport.Outbound {
	return &mocks.Outbound{}
}

func (*mockTransport) Close() error {
	return nil
}
