package mocks

import "github.com/stretchr/testify/mock"

type Logger struct {
	mock.Mock
}

// Debugf provides a mock function with given fields: format, args
func (_m *Logger) Debugf(format string, args ...interface{}) {
	_m.Called(format, args)
}

// Infof provides a mock function with given fields: format, args
func (_m *Logger) Infof(format string, args ...interface{}) {
	_m.Called(format, args)
}

// Warnf provides a mock function with given fields: format, args
func (_m *Logger) Warnf(format string, args ...interface{}) {
	_m.Called(format, args)
}

// Errorf provides a mock function with given fields: format, args
func (_m *Logger) Errorf(format string, args ...interface{}) {
	_m.Called(format, args)
}

// Fatalf provides a mock function with given fields: format, args
func (_m *Logger) Fatalf(format string, args ...interface{}) {
	_m.Called(format, args)
}

// Panicf provides a mock function with given fields: format, args
func (_m *Logger) Panicf(format string, args ...interface{}) {
	_m.Called(format, args)
}
