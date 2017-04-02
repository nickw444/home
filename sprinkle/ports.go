package main

type GPIOManager struct {
	ports       []int
	initialized bool
	rpio        RPIO
}

func NewGPIOManager(ports []int, rpio RPIO) *GPIOManager {
	return &GPIOManager{
		rpio:        rpio,
		ports:       ports,
		initialized: false,
	}
}

func (mgr *GPIOManager) Init() {
	if mgr.initialized {
		return
	}

	err := mgr.rpio.Open()
	if err != nil {
		log.Panic(err)
	}
	mgr.setPortsLow()

	mgr.initialized = true
	log.Printf("Ports initialized successfully.")
}

func (mgr *GPIOManager) Teardown() {
	log.Printf("Tearing down")

	if mgr.initialized {
		mgr.setPortsLow()
		mgr.rpio.Close()
	}
}

func (mgr *GPIOManager) setPortsLow() {
	for _, port := range mgr.ports {
		mgr.rpio.PinMode(port, Output)
		mgr.setPort(port, false)
	}
}

func (mgr *GPIOManager) setPort(port int, on bool) {
	if !mgr.validPort(port) {
		log.Println("Invalid port passed to setPort")
		return
	}

	if on {
		mgr.rpio.SetPinHigh(port)
	} else {
		mgr.rpio.SetPinLow(port)
	}
}

func (mgr *GPIOManager) validPort(port int) bool {
	for _, p := range mgr.ports {
		if port == p {
			return true
		}
	}
	return false
}
