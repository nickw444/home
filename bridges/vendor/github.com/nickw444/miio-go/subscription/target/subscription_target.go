package target

import (
	"github.com/nickw444/miio-go/subscription/common"
	"github.com/nickw444/miio-go/subscription/subscription"
)

type subscriptionTarget struct {
	subscriptions map[string]common.Subscription
}

func NewTarget() common.SubscriptionTarget {
	return &subscriptionTarget{
		subscriptions: make(map[string]common.Subscription),
	}
}

func (t *subscriptionTarget) HasSubscribers() bool {
	return len(t.subscriptions) > 0
}

func (t *subscriptionTarget) Publish(event interface{}) error {
	for _, sub := range t.subscriptions {
		err := sub.Write(event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *subscriptionTarget) NewSubscription() (common.Subscription, error) {
	sub := subscription.NewSubscription(t)
	t.subscriptions[sub.ID()] = sub
	return sub, nil
}

func (t *subscriptionTarget) RemoveSubscription(s common.Subscription) error {
	delete(t.subscriptions, s.ID())
	return nil
}

func (t *subscriptionTarget) CloseAllSubscriptions() error {
	for _, sub := range t.subscriptions {
		err := t.RemoveSubscription(sub)
		if err != nil {
			return err
		}
		err = sub.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
