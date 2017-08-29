package mocks

import "github.com/stretchr/testify/mock"

type Device struct {
	SubscriptionTarget
	mock.Mock
}

// ID provides a mock function with given fields:
func (_m *Device) ID() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetLabel provides a mock function with given fields:
func (_m *Device) GetLabel() (string, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetLabel provides a mock function with given fields: label
func (_m *Device) SetLabel(label string) error {
	ret := _m.Called(label)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(label)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetPower provides a mock function with given fields:
func (_m *Device) GetPower() (bool, error) {
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

// CachedPower provides a mock function with given fields:
func (_m *Device) CachedPower() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SetPower provides a mock function with given fields: state
func (_m *Device) SetPower(state bool) error {
	ret := _m.Called(state)

	var r0 error
	if rf, ok := ret.Get(0).(func(bool) error); ok {
		r0 = rf(state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFirmwareVersion provides a mock function with given fields:
func (_m *Device) GetFirmwareVersion() (string, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CachedFirmwareVersion provides a mock function with given fields:
func (_m *Device) CachedFirmwareVersion() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
