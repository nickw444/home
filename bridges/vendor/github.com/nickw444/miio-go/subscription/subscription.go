package subscription

import (
	"github.com/nickw444/miio-go/subscription/common"
	"github.com/nickw444/miio-go/subscription/target"
)

func NewTarget() common.SubscriptionTarget {
	return target.NewTarget()
}

type SubscriptionTarget = common.SubscriptionTarget
type Subscription = common.Subscription
