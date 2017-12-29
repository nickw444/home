package common

import (
	"sync"
)

// SubscriptionProvider provides an embedable subscription factory
type SubscriptionProvider struct {
	subscriptions map[string]*Subscription
	sync.RWMutex
}

// Subscribe returns a new Subscription for this provider
func (s *SubscriptionProvider) Subscribe() *Subscription {
	s.Lock()
	defer s.Unlock()
	if s.subscriptions == nil {
		s.subscriptions = make(map[string]*Subscription)
	}
	sub := newSubscription(s)
	s.subscriptions[sub.id.String()] = sub

	return sub
}

// Notify sends the provided event to all subscribers
func (s *SubscriptionProvider) Notify(event interface{}) {
	s.RLock()
	defer s.RUnlock()
	for _, sub := range s.subscriptions {
		if err := sub.notify(event); err != nil {
			Log.Warnf("Failed notifying subscription (%s): %s", sub.id, err)
		}
	}
}

// Close all subscriptions
func (s *SubscriptionProvider) Close() (err error) {
	for _, sub := range s.subscriptions {
		serr := sub.Close()
		if serr != nil {
			err = serr
		}
	}

	return err
}

func (s *SubscriptionProvider) unsubscribe(sub *Subscription) error {
	s.Lock()
	defer s.Unlock()
	id := sub.id.String()
	if _, ok := s.subscriptions[id]; !ok {
		return ErrNotFound
	}
	if s.subscriptions != nil {
		delete(s.subscriptions, id)
	}

	return nil
}
