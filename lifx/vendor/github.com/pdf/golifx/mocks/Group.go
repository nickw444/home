package mocks

import "github.com/pdf/golifx/common"
import "github.com/stretchr/testify/mock"

import "time"

type Group struct {
	SubscriptionTarget
	mock.Mock
}

// ID provides a mock function with given fields:
func (_m *Group) ID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetLabel provides a mock function with given fields:
func (_m *Group) GetLabel() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Devices provides a mock function with given fields:
func (_m *Group) Devices() []common.Device {
	ret := _m.Called()

	var r0 []common.Device
	if rf, ok := ret.Get(0).(func() []common.Device); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]common.Device)
		}
	}

	return r0
}

// Lights provides a mock function with given fields:
func (_m *Group) Lights() []common.Light {
	ret := _m.Called()

	var r0 []common.Light
	if rf, ok := ret.Get(0).(func() []common.Light); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]common.Light)
		}
	}

	return r0
}

// GetPower provides a mock function with given fields:
func (_m *Group) GetPower() (bool, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetColor provides a mock function with given fields:
func (_m *Group) GetColor() (common.Color, error) {
	ret := _m.Called()

	var r0 common.Color
	if rf, ok := ret.Get(0).(func() common.Color); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(common.Color)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetColor provides a mock function with given fields: color, duration
func (_m *Group) SetColor(color common.Color, duration time.Duration) error {
	ret := _m.Called(color, duration)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.Color, time.Duration) error); ok {
		r0 = rf(color, duration)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPower provides a mock function with given fields: state
func (_m *Group) SetPower(state bool) error {
	ret := _m.Called(state)

	var r0 error
	if rf, ok := ret.Get(0).(func(bool) error); ok {
		r0 = rf(state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPowerDuration provides a mock function with given fields: state, duration
func (_m *Group) SetPowerDuration(state bool, duration time.Duration) error {
	ret := _m.Called(state, duration)

	var r0 error
	if rf, ok := ret.Get(0).(func(bool, time.Duration) error); ok {
		r0 = rf(state, duration)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
