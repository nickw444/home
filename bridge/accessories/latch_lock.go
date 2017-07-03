package accessories

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/nickw444/homekit/bridge/mqtt"
)

type LatchLockConfig struct {
	LatchDelay int
}

type LatchLock struct {
	domain    *mqtt.Domain
	accessory *accessory.Accessory
	lockSvc   *service.LockMechanism
	log       *logrus.Entry
	conf      LatchLockConfig
}

func NewLatchLockConfig(c map[string]interface{}) LatchLockConfig {
	conf := LatchLockConfig{}
	conf.LatchDelay = 5000

	if val, ok := c["latch-delay"]; ok {
		conf.LatchDelay = val.(int)
	}

	return conf
}

func NewLatchLock(lockConfig LatchLockConfig, client mqtt.Client, identifier, name string, log *logrus.Entry) *LatchLock {
	acc := accessory.New(accessory.Info{
		SerialNumber: identifier,
		Name:         name,
		Model:        "mqtt-latchlock",
	}, accessory.TypeDoorLock)

	lockSvc := service.NewLockMechanism()
	acc.AddService(lockSvc.Service)

	topicSvc := mqtt.NewPrefixedIDTopicService("esp", identifier)
	lock := &LatchLock{
		domain:    mqtt.NewDomain(client, topicSvc),
		accessory: acc,
		lockSvc:   lockSvc,
		log:       log,
		conf:      lockConfig,
	}

	lockSvc.LockTargetState.OnValueRemoteUpdate(func(target int) {
		if target == characteristic.LockTargetStateUnsecured {
			log.Infof("Triggering for %dms\n", lock.conf.LatchDelay)
			lock.domain.Publish("trigger", fmt.Sprintf("T%d", lock.conf.LatchDelay))
		}
	})

	// Subscribe to the door state changing
	lock.domain.Subscribe("relayState", lock.handleLockStatusChange)

	lock.domain.Republish()

	return lock
}

func (l *LatchLock) handleLockStatusChange(msg string) {
	l.log.Infof("Lock Status Changed to %s\n", msg)
	status, err := NewRelayState(msg)
	if err != nil {
		l.log.Error(err)
		return
	}

	if status == relayOpen {
		l.lockSvc.LockCurrentState.SetValue(characteristic.LockCurrentStateUnsecured)
		l.lockSvc.LockTargetState.SetValue(characteristic.LockTargetStateUnsecured)
	} else if status == relayClosed {
		l.lockSvc.LockCurrentState.SetValue(characteristic.LockCurrentStateSecured)
		l.lockSvc.LockTargetState.SetValue(characteristic.LockTargetStateSecured)
	}
}

// GetHCAccessory returns the homecontrol accessory.
func (l *LatchLock) GetHCAccessory() *accessory.Accessory {
	return l.accessory
}

// RelayState represents a concrete type for relay status
type RelayState int

const (
	relayOpen RelayState = iota
	relayClosed
	relayUnknown
)

func NewRelayState(val string) (status RelayState, err error) {
	if val == "OPEN" {
		status = relayOpen
	} else if val == "CLOSED" {
		status = relayClosed
	} else if val == "UNKNOWN" {
		status = relayUnknown
	} else {
		err = fmt.Errorf("Unknown relay status %s", val)
	}
	return
}
