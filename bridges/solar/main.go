package main

import (
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/nickw444/homekit/bridges/lib/environment"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var log = logrus.New()

func main() {
	app := kingpin.New("Efergy Bridge", "Bridge Efergy API with Homekit")
	env := environment.New(app).
		WithDebug().
		WithHC()
	var (
		accessToken = app.Arg("efergy-access-token", "Access token used to connect to efergy API").
				Required().String()
		refreshInt = app.Flag("refresh-interval", "Efergy API Refresh Interval").
				Default("60").Short('r').Int()
	)

	env.Parse()
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

	t, err := hc.NewIPTransport(hc.Config{Pin: env.HC.AccessCode, Port: env.HC.Port}, a)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
		timer.Stop()
	})

	t.Start()
}
