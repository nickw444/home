package mqtt

import (
	"github.com/Sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MessageHandler func(msg string)

type Client interface {
	Publish(topic string, qos byte, retained bool, payload interface{})
	Subscribe(topic string, qos byte, callback MessageHandler)
}

type LoggingClient struct {
	Log *logrus.Entry
}

func (l *LoggingClient) Publish(topic string, qos byte, retained bool, payload interface{}) {
	l.Log.Infof("Publish to topic %s, data: %s", topic, payload)
}
func (l *LoggingClient) Subscribe(topic string, qos byte, callback MessageHandler) {
	l.Log.Infof("Subscribed to topic %s", topic)
}

type PahoClient struct {
	Client mqtt.Client
}

func (p *PahoClient) Publish(topic string, qos byte, retained bool, payload interface{}) {
	p.Client.Publish(topic, qos, retained, payload)
}
func (p *PahoClient) Subscribe(topic string, qos byte, callback MessageHandler) {
	p.Client.Subscribe(topic, qos, func(c mqtt.Client, msg mqtt.Message) {
		callback(string(msg.Payload()))
	})
}
