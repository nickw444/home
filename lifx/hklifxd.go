package main

import (
	"flag"
	"math"
	"os"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"

	"github.com/pdf/golifx"
	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol"
	"github.com/pdf/golifx/protocol/v2/shared"
)

const (
	// from https://github.com/LIFX/LIFXKit/blob/master/LIFXKit/Classes-Common/LFXHSBKColor.h
	HSBKKelvinDefault = uint16(3500)
	HSBKKelvinMin     = uint16(2500)
	HSBKKelvinMax     = uint16(9000)

	HueMax        float64 = 360
	BrightnessMax int     = 100
	SaturationMax float64 = 100
)

type HKLight struct {
	accessory *accessory.Lightbulb
	transport hc.Transport
	sub       *common.Subscription
}

var (
	lights             map[uint64]*HKLight
	pin                string
	transitionDuration time.Duration
)

func Connect() {
	client, err := golifx.NewClient(&protocol.V2{Reliable: true})
	if err != nil {
		log.Info.Panic("Failed to initiliaze the client: %s", err)
	}

	client.SetDiscoveryInterval(30 * time.Second)

	sub, _ := client.NewSubscription()
	for {
		event := <-sub.Events()
		switch event.(type) {
		case common.EventNewLocation:
			log.Debug.Printf("Discovered Location %s", event.(common.EventNewLocation).Location.GetLabel())
		case common.EventNewGroup:
			log.Debug.Printf("Discovered Group %s", event.(common.EventNewGroup).Group.GetLabel())
		case common.EventNewDevice:
			label, _ := event.(common.EventNewDevice).Device.GetLabel()
			log.Debug.Printf("Discovered Device %s", label)

			go NewDevice(event.(common.EventNewDevice).Device)

		case common.EventExpiredLocation:
			log.Debug.Printf("Expired Location %s", event.(common.EventExpiredLocation).Location.GetLabel())
		case common.EventExpiredGroup:
			log.Debug.Printf("Expired Group %s", event.(common.EventExpiredGroup).Group.GetLabel())
		case common.EventExpiredDevice:
			label, _ := event.(common.EventExpiredDevice).Device.GetLabel()
			log.Debug.Printf("Expired Device %s", label)

			ExpireDevice(event.(common.EventExpiredDevice).Device)

		default:
			log.Debug.Printf("Unknown Client Event: %T", event)
		}
	}
}

func NewDevice(device common.Device) {
	if light, ok := device.(common.Light); ok {
		hkLight := GetHKLight(light)

		hkLight.sub, _ = light.NewSubscription()
		for {
			event := <-hkLight.sub.Events()
			switch event.(type) {
			case common.EventUpdateLabel:
				log.Debug.Printf("Updated Label for %s to %s", hkLight.accessory.Info.Name.GetValue(), event.(common.EventUpdateLabel).Label)
				// TODO Add support for label changes to HomeControl
				log.Debug.Printf("Unsupported by HomeControl")
			case common.EventUpdatePower:
				log.Debug.Printf("Updated Power for %s", hkLight.accessory.Info.Name.GetValue())
				hkLight.accessory.Lightbulb.On.SetValue(event.(common.EventUpdatePower).Power)
			case common.EventUpdateColor:
				log.Debug.Printf("Updated Color for %s", hkLight.accessory.Info.Name.GetValue())

				hue, saturation, brightness := ConvertLIFXColor(event.(common.EventUpdateColor).Color)

				hkLight.accessory.Lightbulb.Hue.SetValue(hue)
				hkLight.accessory.Lightbulb.Saturation.SetValue(saturation)
				hkLight.accessory.Lightbulb.Brightness.SetValue(int(brightness))
			case shared.EventBroadcastSent:
				// Suppress event

			default:
				log.Debug.Printf("Unknown Device Event: %T", event)
			}
		}
	} else {
		log.Info.Println("Unsupported Device")
	}
}

func ExpireDevice(device common.Device) {
	if light, ok := device.(common.Light); ok {
		if hkLight, found := lights[light.ID()]; found == true {
			light.CloseSubscription(hkLight.sub)
			hkLight.transport.Stop()

			delete(lights, light.ID())
		}
	} else {
		log.Info.Println("Unsupported Device")
	}
}

func GetHKLight(light common.Light) *HKLight {
	hkLight, found := lights[light.ID()]
	if found {
		return hkLight
	}

	label, _ := light.GetLabel()
	log.Debug.Printf("Creating New HKLight for %s", label)

	info := accessory.Info{
		Name:         label,
		Manufacturer: "LIFX",
	}

	acc := accessory.NewLightbulb(info)

	power, _ := light.GetPower()
	acc.Lightbulb.On.SetValue(power)

	color, _ := light.GetColor()
	hue, saturation, brightness := ConvertLIFXColor(color)

	acc.Lightbulb.Brightness.SetValue(int(brightness))
	acc.Lightbulb.Saturation.SetValue(saturation)
	acc.Lightbulb.Hue.SetValue(hue)

	config := hc.Config{Pin: pin}
	transport, err := hc.NewIPTransport(config, acc.Accessory)
	if err != nil {
		log.Info.Panic(err)
	}

	go func() {
		transport.Start()
	}()

	hkLight = &HKLight{acc, transport, nil}
	lights[light.ID()] = hkLight

	acc.OnIdentify(func() {
		timeout := 1 * time.Second

		for i := 0; i < 4; i++ {
			ToggleLight(light)
			time.Sleep(timeout)
		}
	})

	acc.Lightbulb.On.OnValueRemoteUpdate(func(power bool) {
		log.Debug.Printf("Changed State for %s", label)
		light.SetPowerDuration(power, transitionDuration)
	})

	updateColor := func(light common.Light) {
		currentPower, _ := light.GetPower()

		// HAP: [0...360]
		// LIFX: [0...MAX_UINT16]
		hue := acc.Lightbulb.Hue.GetValue()

		// HAP: [0...100]
		// LIFX: [0...MAX_UINT16]
		saturation := acc.Lightbulb.Saturation.GetValue()

		// HAP: [0...100]
		// LIFX: [0...MAX_UINT16]
		brightness := acc.Lightbulb.Brightness.GetValue()

		// [HSBKKelvinMin..HSBKKelvinMax]
		kelvin := HSBKKelvinDefault

		lifxHue := math.MaxUint16 * float64(hue) / float64(HueMax)
		lifxSaturation := math.MaxUint16 * float64(saturation) / float64(SaturationMax)
		lifxBrightness := math.MaxUint16 * float64(brightness) / float64(BrightnessMax)

		color := common.Color{
			uint16(lifxHue),
			uint16(lifxSaturation),
			uint16(lifxBrightness),
			kelvin,
		}

		light.SetColor(color, transitionDuration)

		if brightness > 0 && !currentPower {
			log.Debug.Printf("Color changed for %s, turning on power.", label)
			light.SetPowerDuration(true, transitionDuration)
		} else if brightness == 0 && currentPower {
			log.Debug.Printf("Color changed for %s, but brightness = 0 turning off power.", label)
			light.SetPower(false)
		}
	}

	acc.Lightbulb.Hue.OnValueRemoteUpdate(func(value float64) {
		log.Debug.Printf("Changed Hue for %s to %f", label, value)
		updateColor(light)
	})

	acc.Lightbulb.Saturation.OnValueRemoteUpdate(func(value float64) {
		log.Debug.Printf("Changed Saturation for %s to %f", label, value)
		updateColor(light)
	})

	acc.Lightbulb.Brightness.OnValueRemoteUpdate(func(value int) {
		log.Debug.Printf("Changed Brightness for %s to %d", label, value)
		updateColor(light)
	})

	return hkLight
}

func Round(f float64) float64 {
	return math.Floor(f + 0.5)
}

func ConvertLIFXColor(color common.Color) (float64, float64, float64) {
	hue := float64(color.Hue) / float64(math.MaxUint16) * float64(HueMax)
	saturation := float64(color.Saturation) / float64(math.MaxUint16) * float64(SaturationMax)
	brightness := float64(color.Brightness) / float64(math.MaxUint16) * float64(BrightnessMax)

	// Return rounded value, HomeKit only uses integer value
	return Round(hue), Round(saturation), Round(brightness)
}

func ToggleLight(light common.Light) {
	power, _ := light.GetPower()
	light.SetPower(!power)
}

func main() {
	lights = map[uint64]*HKLight{}

	pinArg := flag.String("pin", "", "PIN used to pair the LIFX bulbs with HomeKit")
	verboseArg := flag.Bool("v", false, "Whether or not log output is displayed")
	transitionArg := flag.Float64("transition-duration", 1, "Transition time in seconds")

	flag.Parse()

	pin = *pinArg

	if *verboseArg {
		log.Debug.Enable()
	}

	transitionDuration = time.Duration(*transitionArg) * time.Second

	hc.OnTermination(func() {
		for _, light := range lights {
			light.transport.Stop()
		}

		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	})

	Connect()
}
