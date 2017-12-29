package shared

// EventBroadcastSent is emitted when a request is sent via broadcast, to allow
// synchronization of rate limiting
type EventBroadcastSent struct{}

// EventRequestSent is emitted when a request is sent by a device, to allow
// synchronization of rate limiting
type EventRequestSent struct{}
