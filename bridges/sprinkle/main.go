package main

import (
	"os"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hcLog "github.com/brutella/hc/log"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var log = logrus.New()

func makeBridge(config *Config, gpioMgr *GPIOManager) *accessory.Accessory {
	info := accessory.Info{
		Name:         config.Bridge.Name,
		Manufacturer: config.Manufacturer,
		SerialNumber: config.Bridge.Serial,
		Model:        "sprinkle",
	}
	acc := accessory.New(info, accessory.TypeSwitch)

	for _, c := range config.Circuits {
		log.Printf("Registered Circuit '%s' with port %d", c.Name, c.BcmPort)
		circ := NewCircuit(c.Name, c.BcmPort, c.MaxDuration, gpioMgr)
		acc.AddService(circ.Service)
	}

	return acc
}

func main() {

	var (
		app        = kingpin.New("sprinkle", "Homekit Sprinkler Control")
		configFile = app.Flag("config", "Provide a configuration file.").Default("sprinkle.conf.yml").String()
		debug      = app.Flag("debug", "Enable debug output").Bool()
		hcDebug    = app.Flag("hc-debug", "Enable debug output for hc library").Bool()

		run         = app.Command("run", "Run the HAP Server.").Default()
		accessCode  = run.Arg("accessCode", "Homekit Access code to use").Required().String()
		port        = run.Arg("port", "Port for homekit to listen on.").String()
		withoutGPIO = run.Flag("without-gpio", "Disable GPIO. Useful for devlepment on a platform without GPIO").Bool()

		sampleGen   = app.Command("sample-config", "Generate a sample config file")
		numCircuits = sampleGen.Arg("num-circuits", "Number of circuits to generate in the sample config.").Required().Int()
	)

	app.PreAction(func(ctx *kingpin.ParseContext) error {
		if *debug {
			log.Level = logrus.DebugLevel
		}

		if *hcDebug {
			hcLog.Debug.Enable()
			hcLog.Info.Enable()
		}
		return nil
	})

	run.Action(func(ctx *kingpin.ParseContext) error {
		config := ParseConfig(*configFile)
		var rpio RPIO
		if *withoutGPIO {
			rpio = NewFakeRPIO()
		} else {
			rpio = NewHardwareRPIO()
		}

		gpioMgr := NewGPIOManager(config.GetUsedPorts(), rpio)

		defer gpioMgr.Teardown()
		gpioMgr.Init()

		// Make the Bridge and accessories.
		bridge := makeBridge(config, gpioMgr)

		t, err := hc.NewIPTransport(hc.Config{Pin: *accessCode, Port: *port}, bridge)
		if err != nil {
			log.Panic(err)
		}

		hc.OnTermination(func() {
			t.Stop()
		})

		log.Printf("Started server with access code: %s", *accessCode)
		t.Start()
		return nil
	})
	sampleGen.Action(func(ctx *kingpin.ParseContext) error {
		GenerateConfig(*configFile, *numCircuits)
		return nil
	})

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
