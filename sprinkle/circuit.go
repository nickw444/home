package main

import (
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"time"
)

type Circuit struct {
	*service.Switch
	gpioMgr       *GPIOManager
	isIdentifying bool
	port          int
	maxDuration   int
	offTimer      *time.Timer
}

func NewCircuit(name string, port int, maxDuration int, gpio *GPIOManager) *Circuit {

	svc := service.NewSwitch()

	nameChar := characteristic.NewName()
	nameChar.SetValue(name)
	svc.AddCharacteristic(nameChar.Characteristic)

	circ := &Circuit{
		Switch:        svc,
		isIdentifying: false,
		gpioMgr:       gpio,
		port:          port,
		maxDuration:   maxDuration,
	}

	svc.On.OnValueRemoteUpdate(circ.OnUpdate)
	return circ
}

func (c *Circuit) Identify() {
	if !c.isIdentifying {
		c.isIdentifying = true
		go func() {
			c.gpioMgr.setPort(c.port, true)
			time.Sleep(time.Second * 2)
			c.gpioMgr.setPort(c.port, false)
			time.Sleep(time.Second * 1)
			c.gpioMgr.setPort(c.port, true)
			time.Sleep(time.Second * 2)
			c.gpioMgr.setPort(c.port, false)
			c.isIdentifying = false
		}()
	} else {
		log.Printf("Not triggering identification for port %d since it's "+
			"already identifying", c.port)
	}
}

func (c *Circuit) OnUpdate(on bool) {
	// Remove any existing timers.
	if c.offTimer != nil {
		c.offTimer.Stop()
		c.offTimer = nil
	}

	// Turn the port on,
	c.gpioMgr.setPort(c.port, on)
	if on && c.maxDuration > 0 {
		log.Printf("Configuring expiry for port %d", c.port)
		// Turn the port off after the maximum duration
		cb := func() {
			log.Printf("Duration expired for port %d", c.port)
			c.Switch.On.SetValue(false)
			c.gpioMgr.setPort(c.port, false)
		}
		c.offTimer = time.AfterFunc(time.Minute*time.Duration(c.maxDuration), cb)
	}

}
