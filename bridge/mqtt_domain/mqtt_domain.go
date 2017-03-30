package mqtt_domain

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	topic_service "github.com/nickw444/homekit/bridge/topic_service"

	"fmt"
)

const (
	topicEndpointReboot    = "reboot"
	topicEndpointRepublish = "republish"
	topicEndpointReset     = "reset"
)

type MQTTDomain struct {
	client       mqtt.Client
	topicService topic_service.TopicService
}

func NewMQTTDomain(client mqtt.Client, topicService topic_service.TopicService) *MQTTDomain {

	return &MQTTDomain{
		client:       client,
		topicService: topicService,
	}
}

func (s *MQTTDomain) Republish() {
	topic := s.topicService.GetTopic(topicEndpointRepublish)
	fmt.Println(topic)
	s.client.Publish(topic, 0, false, "")
}

func (s *MQTTDomain) Reset() {
	topic := s.topicService.GetTopic(topicEndpointReset)
	s.client.Publish(topic, 0, false, "")
}

func (s *MQTTDomain) Reboot() {
	topic := s.topicService.GetTopic(topicEndpointReboot)
	s.client.Publish(topic, 0, false, "")
}

func (s *MQTTDomain) Publish(topic string, payload interface{}) mqtt.Token {
	fullTopic := s.topicService.GetTopic(topic)
	return s.client.Publish(fullTopic, 0, false, payload)
}

func (s *MQTTDomain) Subscribe(topic string, callback mqtt.MessageHandler) mqtt.Token {
	fullTopic := s.topicService.GetTopic(topic)
	return s.client.Subscribe(fullTopic, 0, callback)
}
