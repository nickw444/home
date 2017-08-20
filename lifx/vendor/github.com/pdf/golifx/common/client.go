package common

import "time"

// Client defines the interface required by protocols
type Client interface {
	GetTimeout() *time.Duration
	GetRetryInterval() *time.Duration
}
