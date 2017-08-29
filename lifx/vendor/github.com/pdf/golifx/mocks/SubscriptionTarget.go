package mocks

import "github.com/pdf/golifx/common"
import "github.com/stretchr/testify/mock"

type SubscriptionTarget struct {
	mock.Mock
}

// NewSubscription provides a mock function with given fields:
func (_m *SubscriptionTarget) NewSubscription() (*common.Subscription, error) {
	ret := _m.Called()

	var r0 *common.Subscription
	if rf, ok := ret.Get(0).(func() *common.Subscription); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*common.Subscription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CloseSubscription provides a mock function with given fields: _a0
func (_m *SubscriptionTarget) CloseSubscription(_a0 *common.Subscription) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*common.Subscription) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
