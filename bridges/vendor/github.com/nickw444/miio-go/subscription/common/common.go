package common

type Subscription interface {
	ID() string
	Events() <-chan interface{}
	Write(event interface{}) error
	Close() error
}

type SubscriptionTarget interface {
	HasSubscribers() bool
	Publish(event interface{}) error
	NewSubscription() (Subscription, error)
	RemoveSubscription(s Subscription) error
	CloseAllSubscriptions() error
}
