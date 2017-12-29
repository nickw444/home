package common

import (
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

const subscriptionChanSize = 16

// SubscriptionTarget generally embeds a SubscriptionProvider
type SubscriptionTarget interface {
	Subscribe() *Subscription
	Notify(event interface{})
}

// Subscription exposes an event channel for consumers, and attaches to a
// SubscriptionTarget, that will feed it with events
type Subscription struct {
	events   chan interface{}
	quitChan chan struct{}
	id       uuid.UUID
	provider *SubscriptionProvider
	sync.Mutex
}

// Events returns a chan reader for reading events published to this
// subscription
func (s *Subscription) Events() <-chan interface{} {
	return s.events
}

// notify pushes an event onto the events channel
func (s *Subscription) notify(event interface{}) error {
	timeout := time.After(DefaultTimeout)
	select {
	case <-s.quitChan:
		Log.Debugf("Subscription %s already closed", s.id)
		return ErrClosed
	case s.events <- event:
		return nil
	case <-timeout:
		Log.Debugf("Timeout on subscription %s", s.id)
		return ErrTimeout
	}
}

// Close cleans up resources and notifies the provider that the subscription
// should no longer be used.  It is important to close subscriptions when you
// are done with them to avoid blocking operations.
func (s *Subscription) Close() error {
	s.Lock()
	defer s.Unlock()
	select {
	case <-s.quitChan:
		Log.Debugf("Subscription %s already closed", s.id)
		return ErrClosed
	default:
		close(s.quitChan)
		close(s.events)
	}
	return s.provider.unsubscribe(s)
}

// newSubscription instantiates a new Subscription
func newSubscription(provider *SubscriptionProvider) *Subscription {
	return &Subscription{
		id:       uuid.NewV4(),
		events:   make(chan interface{}, subscriptionChanSize),
		quitChan: make(chan struct{}),
		provider: provider,
	}
}
