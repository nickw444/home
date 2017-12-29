package mqtt

import (
	"crypto/tls"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
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
	log    *logrus.Entry
}

func NewPahoClient(broker string, user string, password string, tlsSkipVerify bool, log *logrus.Entry) *PahoClient {
	clientOptions := mqtt.NewClientOptions().AddBroker(broker)
	if user != "" {
		clientOptions = clientOptions.SetUsername(user)
	}
	if password != "" {
		clientOptions = clientOptions.SetPassword(password)
	}
	if tlsSkipVerify {
		clientOptions = clientOptions.SetTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}

	pahoClient := mqtt.NewClient(clientOptions)

	return &PahoClient{
		Client: pahoClient,
		log:    log,
	}
}

func (p *PahoClient) Connect() error {
	p.log.Debugf("Attempting connection...")
	if token := p.Client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	p.log.Debugf("Connection successful.")
	return nil
}

func (p *PahoClient) Publish(topic string, qos byte, retained bool, payload interface{}) {
	p.log.Debugf("Publish to topic %s. Payload: %s", topic, payload)
	p.Client.Publish(topic, qos, retained, payload)
}

func (p *PahoClient) Subscribe(topic string, qos byte, callback MessageHandler) {
	p.log.Debugf("Subscribe to topic %s", topic)
	p.Client.Subscribe(topic, qos, func(c mqtt.Client, msg mqtt.Message) {
		callback(string(msg.Payload()))
	})
}
