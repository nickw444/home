package main

import (
	"fmt"
	"time"
)

type RaexBlindClient struct {
	remote  int
	channel int

	rfService    *RFService
	numRepeats   int
	sendDelay    time.Duration
	sendQuitChan chan struct{}
}

type raexBlindAction int

const (
	raexBlindActionUp   raexBlindAction = 254
	raexBlindActionHold                 = 253
	raexBlindActionDown                 = 252
	raexBlindActionPair                 = 127
)

func NewRaexBlindClient(remote int, channel int, numRepeats int, sendDelay time.Duration, rfService *RFService) *RaexBlindClient {
	return &RaexBlindClient{
		remote:     remote,
		channel:    channel,
		rfService:  rfService,
		numRepeats: numRepeats,
		sendDelay:  sendDelay,
	}
}

func (r *RaexBlindClient) Hold() {
	r.sendAction(raexBlindActionHold)
}

func (r *RaexBlindClient) Down() {
	r.sendAction(raexBlindActionDown)
}

func (r *RaexBlindClient) Up() {
	r.sendAction(raexBlindActionUp)
}

func (r *RaexBlindClient) Pair() {
	r.sendAction(raexBlindActionPair)
}

func (r *RaexBlindClient) sendAction(action raexBlindAction) {
	if r.sendQuitChan != nil {
		close(r.sendQuitChan)
	}
	r.sendQuitChan = make(chan struct{})
	go r.sendActionAsync(action)
}

// RAEX blinds appear to go to sleep after periods of inactivity. By sending
// the message multiple times, we can avoid this.
func (r *RaexBlindClient) sendActionAsync(action raexBlindAction) {
	payload := r.makePayload(action)
	for i := 0; i < r.numRepeats; i++ {
		r.rfService.Transmit("raex", payload)
		select {
		case <-r.sendQuitChan:
			return
		case <-time.After(r.sendDelay):
			continue
		}
	}
}

func (r *RaexBlindClient) makePayload(action raexBlindAction) string {
	return fmt.Sprintf("%d:%d:%d:", r.channel, r.remote, action)
}
