package main

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"gopkg.in/alecthomas/kingpin.v2"
)

var log = logrus.New()

func main() {
	var (
		app = kingpin.New("Efergy Bridge", "Bridge Efergy API with Homekit")

		accessCode  = app.Arg("access-code", "Homekit Access code to use").Required().String()
		accessToken = app.Arg("efergy-access-token", "Access token used to connect to efergy API").
				Required().String()

		port       = app.Flag("port", "Port for homekit to listen on").Short('p').String()
		refreshInt = app.Flag("refresh-interval", "Efergy API Refresh Interval").
				Default("60").Short('r').Int()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	efergyClient := NewEfergyClient(*accessToken, log.WithField("component", "efergy-client"))

	info := accessory.Info{
		Name:         "EfergyBridge",
		Manufacturer: "Nick Whyte",
		Model:        "PowerMeter",
		SerialNumber: "000001",
	}
	a := accessory.New(info, accessory.TypeOther)
	svc := NewPowerService("Energy", efergyClient, log.WithField("component", "hc-service"))
	a.AddService(svc.Service)

	var timer *time.Timer
	refresh := time.Second * time.Duration(*refreshInt)
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
