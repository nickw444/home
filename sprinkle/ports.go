package main

import (
	"github.com/stianeikeland/go-rpio"
)

type GPIOManager interface {
	Init()
	Teardown()
	setPortsLow()
	setPort(port int, on bool)
}

type BCMGPIOManager struct {
	ports       []int
	initialized bool
}

func NewBCMGPIOManager(ports []int) *BCMGPIOManager {
	return &BCMGPIOManager{
		ports:       ports,
		initialized: false,
	}
}

func (mgr *BCMGPIOManager) Init() {
	if mgr.initialized {
		return
	}

	err := rpio.Open()
	if err != nil {
		log.Panic(err)
	}
	mgr.setPortsLow()

	mgr.initialized = true
	log.Printf("Ports initialized successfully.")
}

func (mgr *BCMGPIOManager) Teardown() {
	log.Printf("Tearing down")

	if mgr.initialized {
		mgr.setPortsLow()
		rpio.Close()
	}
}

func (mgr *BCMGPIOManager) setPortsLow() {
	for _, port := range mgr.ports {
		pin := rpio.Pin(port)
		pin.Output()
		mgr.setPort(port, false)
	}
}

func (mgr *BCMGPIOManager) setPort(port int, on bool) {
	log.Printf("Setting port %d to %d", port, on)
	if !mgr.validPort(port) {
		log.Println("Invalid port passed to setPort")
		return
	}

	pin := rpio.Pin(port)
	if on {
		pin.High()
	} else {
		pin.Low()
	}
}

func (mgr *BCMGPIOManager) validPort(port int) bool {
	for _, p := range mgr.ports {
		if port == p {
			return true
		}
	}
	return false
}

// FakeGPIOManager is a fake GPIOManager implementation for use for dev without
// accessible GPIO
type FakeGPIOManager struct {
	bcm   *BCMGPIOManager
	ports []int
}

func NewFakeGPIOManager(ports []int) *FakeGPIOManager {
	return &FakeGPIOManager{
		bcm:   NewBCMGPIOManager(ports),
		ports: ports,
	}
}

func (f *FakeGPIOManager) Init() {
	f.setPortsLow()
	log.Printf("Ports initialized successfully.")
}
func (f *FakeGPIOManager) Teardown() {
	f.setPortsLow()
}
func (f *FakeGPIOManager) setPortsLow() {
	for _, port := range f.ports {
		f.setPort(port, false)
	}
}
func (f *FakeGPIOManager) setPort(port int, on bool) {
	log.Printf("Setting port %d to %d", port, on)
	if !f.bcm.validPort(port) {
		log.Println("Invalid port passed to setPort")
		return
	}
}
