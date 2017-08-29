package mocks

import "github.com/pdf/golifx/common"
import "github.com/stretchr/testify/mock"

import "time"

type Light struct {
	Device
	mock.Mock
}

// SetColor provides a mock function with given fields: color, duration
func (_m *Light) SetColor(color common.Color, duration time.Duration) error {
	ret := _m.Called(color, duration)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.Color, time.Duration) error); ok {
		r0 = rf(color, duration)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetColor provides a mock function with given fields:
func (_m *Light) GetColor() (common.Color, error) {
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

// CachedColor provides a mock function with given fields:
func (_m *Light) CachedColor() common.Color {
	ret := _m.Called()

	var r0 common.Color
	if rf, ok := ret.Get(0).(func() common.Color); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(common.Color)
	}

	return r0
}

// SetPowerDuration provides a mock function with given fields: state, duration
func (_m *Light) SetPowerDuration(state bool, duration time.Duration) error {
	ret := _m.Called(state, duration)

	var r0 error
	if rf, ok := ret.Get(0).(func(bool, time.Duration) error); ok {
		r0 = rf(state, duration)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
