package device

import (
	"math"
	"time"

	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol/v2/packet"
	"github.com/pdf/golifx/protocol/v2/shared"
)

const (
	Get             shared.Message = 101
	SetColor        shared.Message = 102
	State           shared.Message = 107
	LightGetPower   shared.Message = 116
	LightSetPower   shared.Message = 117
	LightStatePower shared.Message = 118
)

type Light struct {
	*Device
	color common.Color
}

type payloadColor struct {
	Reserved uint8
	Color    common.Color
	Duration uint32
}

type payloadPowerDuration struct {
	Level    uint16
	Duration uint32
}

type state struct {
	Color     common.Color
	Reserved0 int16
	Power     uint16
	Label     [32]byte
	Reserved1 uint64
}

func (l *Light) SetState(pkt *packet.Packet) error {
	s := &state{}

	if err := pkt.DecodePayload(s); err != nil {
		return err
	}
	common.Log.Debugf("Got light state (%d): %+v", l.id, s)

	if !common.ColorEqual(s.Color, l.CachedColor()) {
		l.Lock()
		l.color = s.Color
		l.Unlock()
		if err := l.publish(common.EventUpdateColor{Color: l.color}); err != nil {
			return err
		}
	}
	if s.Power > 0 != l.CachedPower() {
		l.Lock()
		l.power = s.Power
		l.Unlock()
		if err := l.publish(common.EventUpdatePower{Power: l.power > 0}); err != nil {
			return err
		}
	}
	newLabel := stripNull(string(s.Label[:]))
	if newLabel != l.CachedLabel() {
		l.Lock()
		l.label = newLabel
		l.Unlock()
		if err := l.publish(common.EventUpdateLabel{Label: l.label}); err != nil {
			return err
		}
	}

	return nil
}

func (l *Light) Get() error {
	pkt := packet.New(l.address, l.requestSocket)
	pkt.SetType(Get)
	req, err := l.Send(pkt, l.reliable, true)
	if err != nil {
		return err
	}

	common.Log.Debugf("Waiting for light state (%d)", l.id)
	pktResponse := <-req
	if pktResponse.Error != nil {
		return pktResponse.Error
	}

	return l.SetState(&pktResponse.Result)
}

func (l *Light) SetColor(color common.Color, duration time.Duration) error {
	if common.ColorEqual(color, l.CachedColor()) {
		return nil
	}

	common.Log.Debugf("Setting color on %d", l.id)
	if duration < shared.RateLimit {
		duration = shared.RateLimit
	}
	p := &payloadColor{
		Color:    color,
		Duration: uint32(duration / time.Millisecond),
	}

	pkt := packet.New(l.address, l.requestSocket)
	pkt.SetType(SetColor)
	if err := pkt.SetPayload(p); err != nil {
		return err
	}
	req, err := l.Send(pkt, l.reliable, false)
	if err != nil {
		return err
	}
	if l.reliable {
		// Wait for ack
		<-req
		common.Log.Debugf("Setting color on %d acknowledged", l.id)
	}

	l.Lock()
	l.color = color
	l.Unlock()
	return l.publish(common.EventUpdateColor{Color: l.color})
}

func (l *Light) GetColor() (common.Color, error) {
	if err := l.Get(); err != nil {
		return common.Color{}, err
	}
	return l.CachedColor(), nil
}

func (l *Light) CachedColor() common.Color {
	l.RLock()
	defer l.RUnlock()
	return l.color
}

func (l *Light) SetPowerDuration(state bool, duration time.Duration) error {
	p := new(payloadPowerDuration)
	if state {
		p.Level = math.MaxUint16
	}
	p.Duration = uint32(duration / time.Millisecond)

	pkt := packet.New(l.address, l.requestSocket)
	pkt.SetType(LightSetPower)
	if err := pkt.SetPayload(p); err != nil {
		return err
	}

	common.Log.Debugf("Setting power state on %d: %v", l.id, state)
	req, err := l.Send(pkt, l.reliable, false)
	if err != nil {
		return err
	}
	if l.reliable {
		// Wait for ack
		<-req
		common.Log.Debugf("Setting power state on %d acknowledged", l.id)
	}

	l.Lock()
	l.power = p.Level
	l.Unlock()
	return l.publish(common.EventUpdatePower{Power: l.power > 0})
}
