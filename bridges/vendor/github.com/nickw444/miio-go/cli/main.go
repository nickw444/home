package main

import (
	"fmt"

	"time"

	"github.com/nickw444/miio-go"
	"github.com/nickw444/miio-go/device"

	"github.com/nickw444/miio-go/common"
	"github.com/nickw444/miio-go/subscription"
	"github.com/sirupsen/logrus"
)

func onNewDevice(dev common.Device) {
	switch dev.(type) {
	case *device.Yeelight:
		fmt.Printf("Found Yeelight :)\n")
	case *device.PowerPlug:
		fmt.Println("Found PowerPlug")
		d := dev.(*device.PowerPlug)
		sub, err := d.NewSubscription()
		if err != nil {
			panic(err)
		}
		fmt.Println("MEMES")
		go watchSubscription(sub)
		go tick(d)

	default:
		fmt.Printf("Unknown device type %T\n", dev)
	}
}

func watchSubscription(sub subscription.Subscription) {
	for event := range sub.Events() {
		fmt.Printf("New Sub Event: %T\n", event)
	}
}

func tick(d *device.PowerPlug) {
	fmt.Println("TICKING")
	currState := false
	for {
		select {
		case <-time.After(time.Second * 5):
			fmt.Println("AFTER 5 Secnods")
			var s common.PowerState
			if currState {
				s = common.PowerStateOn
			} else {
				s = common.PowerStateOff
			}
			currState = !currState
			d.SetPower(s)
		}
	}
}

func main() {
	l := logrus.New()
	l.SetLevel(logrus.InfoLevel)
	common.SetLogger(l)

	client, err := miio.NewClient()
	if err != nil {
		panic(err)
	}

	client.SetDiscoveryInterval(time.Second * 10)

	sub, err := client.NewSubscription()
	if err != nil {
		panic(err)
	}

	for event := range sub.Events() {
		switch event.(type) {
		case common.EventNewDevice:
			dev := event.(common.EventNewDevice).Device
			onNewDevice(dev)
			fmt.Printf("New device event %T\n", dev)
		case common.EventExpiredDevice:
			fmt.Println("Expired device event")
		default:
			fmt.Println("Uknown event")
		}
	}
}
