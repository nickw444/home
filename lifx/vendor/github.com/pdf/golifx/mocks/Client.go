package mocks

import "github.com/stretchr/testify/mock"

import "time"

type Client struct {
	mock.Mock
}

// GetTimeout provides a mock function with given fields:
func (_m *Client) GetTimeout() *time.Duration {
	ret := _m.Called()

	var r0 *time.Duration
	if rf, ok := ret.Get(0).(func() *time.Duration); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*time.Duration)
		}
	}

	return r0
}

// GetRetryInterval provides a mock function with given fields:
func (_m *Client) GetRetryInterval() *time.Duration {
	ret := _m.Called()

	var r0 *time.Duration
	if rf, ok := ret.Get(0).(func() *time.Duration); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*time.Duration)
		}
	}

	return r0
}
