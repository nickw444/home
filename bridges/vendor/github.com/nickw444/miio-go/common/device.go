package common

import "github.com/nickw444/miio-go/subscription"

type Device interface {
	subscription.SubscriptionTarget

	ID() uint32
	GetLabel() (string, error)
}
