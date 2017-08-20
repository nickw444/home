package common

import "time"

const (
	// DefaultTimeout is the default duration after which operations time out
	DefaultTimeout = 2 * time.Second
	// DefaultRetryInterval is the default interval at which operations are
	// retried
	DefaultRetryInterval = 100 * time.Millisecond
)
