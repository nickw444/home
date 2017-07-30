package rf_service

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/nickw444/homekit/bridge/mqtt"
)

type RFServiceConfig struct {
	serial string
}

type RFService struct {
	domain *mqtt.Domain
	log    *logrus.Entry
}

func NewConfig(c map[string]interface{}) (conf RFServiceConfig) {
	if val, ok := c["serial"]; ok {
		conf.serial = val.(string)
	} else {
		panic(fmt.Errorf("Missing configuration key: serial"))
	}

	return
}

func New(config RFServiceConfig, mqttClient mqtt.Client, log *logrus.Entry) *RFService {
	topicSvc := mqtt.NewPrefixedIDTopicService("esp", config.serial)
	domain := mqtt.NewDomain(mqttClient, topicSvc)

	return &RFService{
		domain: domain,
		log:    log,
	}
}

func (r *RFService) Transmit(endpoint string, payload string) {
	r.domain.Publish(endpoint, payload)
}
