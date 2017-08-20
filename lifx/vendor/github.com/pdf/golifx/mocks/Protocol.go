package mocks

import "github.com/pdf/golifx/common"
import "github.com/stretchr/testify/mock"

import "time"

type Protocol struct {
	SubscriptionTarget
	mock.Mock
}

// GetLocations provides a mock function with given fields:
func (_m *Protocol) GetLocations() ([]common.Location, error) {
	ret := _m.Called()

	var r0 []common.Location
	if rf, ok := ret.Get(0).(func() []common.Location); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]common.Location)
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

// GetLocation provides a mock function with given fields: id
func (_m *Protocol) GetLocation(id string) (common.Location, error) {
	ret := _m.Called(id)

	var r0 common.Location
	if rf, ok := ret.Get(0).(func(string) common.Location); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(common.Location)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGroups provides a mock function with given fields:
func (_m *Protocol) GetGroups() ([]common.Group, error) {
	ret := _m.Called()

	var r0 []common.Group
	if rf, ok := ret.Get(0).(func() []common.Group); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]common.Group)
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

// GetGroup provides a mock function with given fields: id
func (_m *Protocol) GetGroup(id string) (common.Group, error) {
	ret := _m.Called(id)

	var r0 common.Group
	if rf, ok := ret.Get(0).(func(string) common.Group); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(common.Group)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDevices provides a mock function with given fields:
func (_m *Protocol) GetDevices() ([]common.Device, error) {
	ret := _m.Called()

	var r0 []common.Device
	if rf, ok := ret.Get(0).(func() []common.Device); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]common.Device)
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

// GetDevice provides a mock function with given fields: id
func (_m *Protocol) GetDevice(id uint64) (common.Device, error) {
	ret := _m.Called(id)

	var r0 common.Device
	if rf, ok := ret.Get(0).(func(uint64) common.Device); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(common.Device)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Discover provides a mock function with given fields:
func (_m *Protocol) Discover() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTimeout provides a mock function with given fields: timeout
func (_m *Protocol) SetTimeout(timeout *time.Duration) {
	_m.Called(timeout)
}

// SetRetryInterval provides a mock function with given fields: retryInterval
func (_m *Protocol) SetRetryInterval(retryInterval *time.Duration) {
	_m.Called(retryInterval)
}

// Close provides a mock function with given fields:
func (_m *Protocol) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPower provides a mock function with given fields: state
func (_m *Protocol) SetPower(state bool) error {
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
func (_m *Protocol) SetPowerDuration(state bool, duration time.Duration) error {
	ret := _m.Called(state, duration)

	var r0 error
	if rf, ok := ret.Get(0).(func(bool, time.Duration) error); ok {
		r0 = rf(state, duration)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetColor provides a mock function with given fields: color, duration
func (_m *Protocol) SetColor(color common.Color, duration time.Duration) error {
	ret := _m.Called(color, duration)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.Color, time.Duration) error); ok {
		r0 = rf(color, duration)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
