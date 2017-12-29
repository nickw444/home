package subscription

import (
	"errors"
	"sync"
	"time"

	"github.com/nickw444/miio-go/subscription/common"
	"github.com/satori/go.uuid"
)

const (
	defaultTimeout = 2 * time.Second
	chanSize       = 16
)

var (
	ErrClosed  = errors.New("Subscription is already closed.")
	ErrTimeout = errors.New("Timed out.")
)

type subscription struct {
	id       uuid.UUID
	wg       sync.WaitGroup
	quitChan chan struct{}
	events   chan interface{}
	target   common.SubscriptionTarget
}

func NewSubscription(target common.SubscriptionTarget) common.Subscription {
	return &subscription{
		id:       uuid.NewV4(),
		events:   make(chan interface{}, chanSize),
		quitChan: make(chan struct{}),
		target:   target,
	}
}

func (s *subscription) ID() string {
	return s.id.String()
}

func (s *subscription) Events() <-chan interface{} {
	return s.events
}

func (s *subscription) Write(event interface{}) error {
	s.wg.Add(1)
	defer s.wg.Done()
	timeout := time.After(defaultTimeout)
	select {
	case <-s.quitChan:
		return ErrClosed
	default:
	}
	select {
	case <-s.quitChan:
		return ErrClosed
	case s.events <- event:
		return nil
	case <-timeout:
		// TODO NW Warnings
		//Log.Debugf("Timeout on subscription %s", s.ID)
		return ErrTimeout
	}
}

func (s *subscription) Close() error {
	select {
	case <-s.quitChan:
		// TODO NW Warnings
		//Log.Warnf("Subscription %s already closed", s.ID)
		return ErrClosed
	default:
		close(s.quitChan)
		s.wg.Wait()
		close(s.events)
	}
	return s.target.RemoveSubscription(s)
}
