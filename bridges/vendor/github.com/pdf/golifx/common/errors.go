package common

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound not found
	ErrNotFound = errors.New(`Not found`)
	// ErrProtocol protocol error
	ErrProtocol = errors.New(`Protocol error`)
	// ErrDuplicate already exists
	ErrDuplicate = errors.New(`Already exists`)
	// ErrInvalidArgument invalid argument
	ErrInvalidArgument = errors.New(`Invalid argument`)
	// ErrClosed connection closed
	ErrClosed = errors.New(`Connection closed`)
	// ErrTimeout timed out
	ErrTimeout = errors.New(`Timed out`)
	// ErrDeviceInvalidType invalid device type
	ErrDeviceInvalidType = errors.New(`Invalid device type`)
	// ErrUnsupported operation is not supported
	ErrUnsupported = errors.New(`Operation not supported`)
)

// ErrNotImplemented not implemented
type ErrNotImplemented struct {
	Method string
}

// Error satisfies the error interface
func (e *ErrNotImplemented) Error() string {
	return fmt.Sprintf("Method '%s' not implemented for this protocol", e.Method)
}
