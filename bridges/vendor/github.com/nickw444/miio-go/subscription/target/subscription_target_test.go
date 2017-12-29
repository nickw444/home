package target

import (
	"testing"

	"github.com/nickw444/miio-go/subscription/common"
	"github.com/nickw444/miio-go/subscription/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func SubscriptionTarget_Setup() (tt struct {
	target *subscriptionTarget
}) {
	tt.target = &subscriptionTarget{
		subscriptions: make(map[string]common.Subscription),
	}
	return
}

func TestSubscriptionTarget_HasSubscribers(t *testing.T) {
	tt := SubscriptionTarget_Setup()
	assert.False(t, tt.target.HasSubscribers())

	s, err := tt.target.NewSubscription()
	assert.NoError(t, err)
	assert.True(t, tt.target.HasSubscribers())

	s.Close()
	assert.NoError(t, err)
	assert.False(t, tt.target.HasSubscribers())
}

// Ensure that an event is published to all subscriptions
func TestSubscriptionTarget_Publish(t *testing.T) {
	tt := SubscriptionTarget_Setup()

	sub1 := &mocks.Subscription{}
	sub2 := &mocks.Subscription{}

	sub1.On("Write", mock.Anything).Return(nil).Once()
	sub2.On("Write", mock.Anything).Return(nil).Once()

	tt.target.subscriptions = map[string]common.Subscription{
		"01": sub1,
		"02": sub2,
	}

	err := tt.target.Publish(struct{}{})
	assert.NoError(t, err)
	sub1.AssertExpectations(t)
	sub2.AssertExpectations(t)
}

// Ensure new subscriptions are tracked
func TestSubscriptionTarget_NewSubscription(t *testing.T) {
	tt := SubscriptionTarget_Setup()

	assert.Len(t, tt.target.subscriptions, 0)
	s, err := tt.target.NewSubscription()
	assert.NoError(t, err)

	assert.Len(t, tt.target.subscriptions, 1)
	assert.Equal(t, s, tt.target.subscriptions[s.ID()])
}

// Ensure subscriptions are correctly removed.
func TestSubscriptionTarget_RemoveSubscription(t *testing.T) {
	tt := SubscriptionTarget_Setup()

	assert.Len(t, tt.target.subscriptions, 0)
	s1, err := tt.target.NewSubscription()
	assert.NoError(t, err)
	assert.Len(t, tt.target.subscriptions, 1)
	s2, err := tt.target.NewSubscription()
	assert.NoError(t, err)
	assert.Len(t, tt.target.subscriptions, 2)

	tt.target.RemoveSubscription(s1)
	assert.Len(t, tt.target.subscriptions, 1)
	tt.target.RemoveSubscription(s2)
	assert.Len(t, tt.target.subscriptions, 0)
}

// Ensure all subscriptions are closed
func TestSubscriptionTarget_CloseAllSubscriptions(t *testing.T) {
	tt := SubscriptionTarget_Setup()

	sub1 := &mocks.Subscription{}
	sub2 := &mocks.Subscription{}
	sub1.On("Write", mock.Anything).Return(nil).Once()
	sub2.On("Write", mock.Anything).Return(nil).Once()
	sub1.On("ID").Return("01").Once()
	sub2.On("ID").Return("02").Once()
	sub1.On("Close").Return(nil).Once()
	sub2.On("Close").Return(nil).Once()
	tt.target.subscriptions = map[string]common.Subscription{
		"01": sub1,
		"02": sub2,
	}

	err := tt.target.CloseAllSubscriptions()
	assert.NoError(t, err)

	assert.Len(t, tt.target.subscriptions, 0)
}
