package device

import (
	"testing"

	"bytes"

	"time"

	"github.com/benbjohnson/clock"
	"github.com/nickw444/miio-go/device/product"
	"github.com/nickw444/miio-go/mocks"
	"github.com/nickw444/miio-go/protocol/packet"
	mocks2 "github.com/nickw444/miio-go/subscription/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BaseDevice_SetUp() (ret struct {
	deviceId  uint32
	clk       *clock.Mock
	subTgt    *mocks2.SubscriptionTarget
	outbound  *mocks.Outbound
	rThrottle *mocks.RefreshThrottle
	device    *baseDevice
}) {
	ret.deviceId = 10
	ret.clk = clock.NewMock()
	ret.subTgt = &mocks2.SubscriptionTarget{}
	ret.outbound = &mocks.Outbound{}
	ret.rThrottle = &mocks.RefreshThrottle{}
	ret.device = &baseDevice{
		SubscriptionTarget: ret.subTgt,
		refreshThrottle:    ret.rThrottle,
		outbound:           ret.outbound,
		id:                 ret.deviceId,
		seen:               ret.clk.Now(),
	}
	ret.device.init()

	// Tick the clock
	ret.clk.Add(time.Second * 5)

	return
}

// Handle an incoming packet via the outbound
func TestBaseDevice_Handle(t *testing.T) {
	tt := BaseDevice_SetUp()

	pkt := packet.New(tt.deviceId, bytes.Repeat([]byte{0xfa}, 16), 0xAAA, bytes.Repeat([]byte{0xCA}, 10))
	tt.outbound.On("Handle", pkt).Return(nil)

	tt.device.Handle(pkt)
	tt.outbound.AssertExpectations(t)
}

// Handle an incoming packet should update seen time to packet decode time.
func TestBaseDevice_Handle2(t *testing.T) {
	tt := BaseDevice_SetUp()
	pkt := packet.New(tt.deviceId, bytes.Repeat([]byte{0xfa}, 16), 0xAAA, bytes.Repeat([]byte{0xCA}, 10))
	pkt.Meta.DecodeTime = tt.clk.Now()

	tt.outbound.On("Handle", pkt).Return(nil)
	tt.device.Handle(pkt)

	assert.EqualValues(t, pkt.Meta.DecodeTime, tt.device.seen)
}

// Closes subscriptions and refreshThrottle on close
func TestBaseDevice_Close(t *testing.T) {
	tt := BaseDevice_SetUp()

	tt.subTgt.On("CloseAllSubscriptions").Return(nil)
	tt.rThrottle.On("Close")
	tt.device.Close()

	tt.subTgt.AssertExpectations(t)
	tt.rThrottle.AssertExpectations(t)
}

// Sets / Gets Provisional value
func TestBaseDevice_Provisional(t *testing.T) {
	tt := BaseDevice_SetUp()

	assert.True(t, tt.device.provisional)
	tt.device.SetProvisional(false)
	assert.False(t, tt.device.provisional)
	tt.device.SetProvisional(true)
	assert.True(t, tt.device.provisional)
}

func BaseDevice_GetProduct_Setup(outbound *mocks.Outbound) {
	outbound.On("CallAndDeserialize", "miIO.info", mock.AnythingOfType("[]string"), mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			resp := args.Get(2).(*InfoResponse)
			resp.Result.Model = "chuangmi.plug.m1"
		})
}

// GetProduct performs a miIO.info via outbound
func TestBaseDevice_GetProduct(t *testing.T) {
	tt := BaseDevice_SetUp()
	BaseDevice_GetProduct_Setup(tt.outbound)

	_, err := tt.device.GetProduct()
	assert.NoError(t, err)
	tt.outbound.AssertExpectations(t)
}

// GetProduct returns the appropriate product
func TestBaseDevice_GetProduct2(t *testing.T) {
	tt := BaseDevice_SetUp()
	BaseDevice_GetProduct_Setup(tt.outbound)

	p, err := tt.device.GetProduct()
	assert.NoError(t, err)
	assert.Equal(t, product.PowerPlug, p)
}

// Discover sends a hello packet via outbound
func TestBaseDevice_Discover(t *testing.T) {
	tt := BaseDevice_SetUp()

	tt.outbound.On("Send", packet.NewHello()).Return(nil)
	tt.device.Discover()
	tt.outbound.AssertExpectations(t)

}

func BaseDevice_NewSubscription_Setup(throttle *mocks.RefreshThrottle, target *mocks2.SubscriptionTarget) {
	throttle.On("Start")
	target.On("NewSubscription").Return(nil, nil)
}

// starts refreshThrottle when a new subscription is created
func TestBaseDevice_NewSubscription(t *testing.T) {
	tt := BaseDevice_SetUp()
	BaseDevice_NewSubscription_Setup(tt.rThrottle, tt.subTgt)

	_, err := tt.device.NewSubscription()
	assert.NoError(t, err)
	tt.rThrottle.AssertExpectations(t)
}

// creates a new subscription
func TestBaseDevice_NewSubscription2(t *testing.T) {
	tt := BaseDevice_SetUp()
	BaseDevice_NewSubscription_Setup(tt.rThrottle, tt.subTgt)

	_, err := tt.device.NewSubscription()
	assert.NoError(t, err)
	tt.subTgt.AssertExpectations(t)
}

// stops refresh throttle if the last subscription is closed
func TestBaseDevice_RemoveSubscription(t *testing.T) {
	tt := BaseDevice_SetUp()

	tt.subTgt.On("RemoveSubscription", mock.Anything).Return(nil).Times(2)
	tt.rThrottle.On("Stop").Once()

	tt.subTgt.On("HasSubscribers").Return(true).Once()
	err := tt.device.RemoveSubscription(nil)
	assert.NoError(t, err)

	tt.subTgt.On("HasSubscribers").Return(false).Once()
	err = tt.device.RemoveSubscription(nil)
	assert.NoError(t, err)

	tt.subTgt.AssertExpectations(t)
	tt.rThrottle.AssertExpectations(t)
}
