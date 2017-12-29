package capability

import (
	"testing"

	"github.com/nickw444/miio-go/common"
	"github.com/nickw444/miio-go/mocks"
	mocks2 "github.com/nickw444/miio-go/subscription/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Power_SetUp() (tt struct {
	power    *Power
	outbound *mocks.Outbound
	target   *mocks2.SubscriptionTarget
}) {
	tt.target = new(mocks2.SubscriptionTarget)
	tt.outbound = new(mocks.Outbound)
	tt.power = NewPower(tt.target, tt.outbound)
	return
}

// Ensure an event is emitted on power state change
func TestPower_Update(t *testing.T) {
	tt := Power_SetUp()

	tt.outbound.On("CallAndDeserialize", mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			resp := args.Get(2).(*PowerResponse)
			resp.Result = []common.PowerState{common.PowerStateOn}
		})
	tt.target.On("Publish", mock.Anything).Return(nil).Once()

	err := tt.power.Update()
	assert.NoError(t, err)
	tt.target.AssertExpectations(t)
}

// Should bubbdle outbound errors
func TestPower_Update2(t *testing.T) {
	tt := Power_SetUp()

	tt.outbound.On("CallAndDeserialize", mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.Anything).
		Return(assert.AnError)

	err := tt.power.Update()
	assert.Error(t, err)
}

// Ensure an event is emitted on SetPower.
func TestPower_SetPower(t *testing.T) {
	tt := Power_SetUp()

	tt.outbound.On("Call", mock.Anything, mock.Anything).Return(nil, nil)
	tt.target.On("Publish", mock.Anything).Return(nil).Once()

	err := tt.power.SetPower(common.PowerStateOn)
	assert.NoError(t, err)
	tt.target.AssertExpectations(t)
}

// Ensure outbound is called on SetPower
func TestPower_SetPower2(t *testing.T) {
	tt := Power_SetUp()

	tt.target.On("Publish", mock.Anything).Return(nil)
	tt.outbound.
		On("Call", "set_power", []string{common.PowerStateOn}).
		Return(nil, nil)

	err := tt.power.SetPower(common.PowerStateOn)
	assert.NoError(t, err)

	tt.outbound.AssertNumberOfCalls(t, "Call", 1)
	tt.outbound.AssertExpectations(t)
}

// Should bubble outbound errors
func TestPower_SetPower3(t *testing.T) {
	tt := Power_SetUp()

	tt.outbound.
		On("Call", "set_power", []string{common.PowerStateOn}).
		Return(nil, assert.AnError)

	err := tt.power.SetPower(common.PowerStateOn)
	assert.Error(t, err)
}
