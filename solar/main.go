package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"log"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {

	accessToken := kingpin.Arg("accessToken", "Access token used to connect to efergy API").Required().String()
	accessCode := kingpin.Arg("accessCode", "Homekit Access code to use").Required().String()
	port := kingpin.Arg("port", "Port for homekit to listen on").String()
	kingpin.Parse()

	efergyClient := NewEfergyClient(*accessToken)

	info := accessory.Info{
		Name:         "EfergyBridge",
		Manufacturer: "Nick Whyte",
		Model:        "PowerMeter",
		SerialNumber: "000001",
	}
	a := accessory.New(info, accessory.TypeOther)
	svc := NewPowerService("Energy", efergyClient)
	a.AddService(svc.Service)

	var timer *time.Timer
	refresh := time.Second * 60
	timer = time.AfterFunc(refresh, func() {
		svc.Update()
		timer.Reset(refresh)
	})

	svc.Update()

	t, err := hc.NewIPTransport(hc.Config{Pin: *accessCode, Port: *port}, a)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
		timer.Stop()
	})

	t.Start()
}
