package main

import (
	"time"

	"github.com/nickw444/homekit/bridges/lib/mqtt"
)

type transmission struct {
	endpoint string
	payload  string
}

type RFService struct {
	domain *mqtt.Domain
	q      chan *transmission
}

func NewRFService(serial string, mqttClient mqtt.Client) *RFService {
	topicSvc := mqtt.NewPrefixedIDTopicService("esp", serial)
	domain := mqtt.NewDomain(mqttClient, topicSvc)

	svc := &RFService{
		domain: domain,
		q:      make(chan *transmission),
	}

	go svc.worker()

	return svc
}

func (r *RFService) Transmit(endpoint string, payload string) {
	r.q <- &transmission{endpoint, payload}
}

func (r *RFService) worker() {
	for transmission := range r.q {
		r.domain.Publish(transmission.endpoint, transmission.payload)
		// Sending transmissions back to back appears to crash the RF transmitter.
		// Until I find time to resolve the issue in firmware, just rate limit.
		time.Sleep(time.Millisecond * 500)
	}
}
