package capability

import (
	"github.com/nickw444/miio-go/common"
	"github.com/nickw444/miio-go/protocol/transport"
	"github.com/nickw444/miio-go/subscription"
)

type Power struct {
	subscriptionTarget subscription.SubscriptionTarget
	outbound           transport.Outbound
	powerState         common.PowerState
}

type PowerResponse struct {
	Result []common.PowerState `json:"result"`
}

func NewPower(target subscription.SubscriptionTarget, transport transport.Outbound) *Power {
	return &Power{
		subscriptionTarget: target,
		outbound:           transport,
		powerState:         common.PowerStateUnknown,
	}
}

func (p *Power) SetPower(state common.PowerState) error {
	_, err := p.outbound.Call("set_power", []string{string(state)})
	if err != nil {
		return err
	}

	// TODO NW: Use the value from the response here.
	p.powerState = state
	p.subscriptionTarget.Publish(common.EventUpdatePower{p.powerState})

	return nil
}

func (p *Power) Update() error {
	resp := PowerResponse{}
	err := p.outbound.CallAndDeserialize("get_prop", []string{"power"}, &resp)
	if err != nil {
		return err
	}

	if resp.Result[0] != p.powerState {
		p.powerState = resp.Result[0]
		p.subscriptionTarget.Publish(common.EventUpdatePower{p.powerState})
	}

	return nil
}
