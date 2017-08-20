package shared

import "time"

const (
	DefaultPort = 56700
	RateLimit   = time.Second / 20 // Max 20 packets per second
)
