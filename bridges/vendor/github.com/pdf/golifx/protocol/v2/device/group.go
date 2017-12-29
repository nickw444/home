package device

import (
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol/v2/packet"
)

type stateGroup struct {
	ID        [16]byte `struc:"little"`
	Label     [32]byte `struc:"little"`
	UpdatedAt uint64   `struc:"little"`
}

type Group struct {
	id        [16]byte
	idEncoded string
	label     [32]byte
	updatedAt uint64
	devices   map[uint64]GenericDevice
	quitChan  chan struct{}
	color     common.Color
	power     bool
	common.SubscriptionProvider
	sync.RWMutex
}

func (g *Group) init() {
	g.Lock()
	g.devices = make(map[uint64]GenericDevice)
	g.quitChan = make(chan struct{})
	g.Unlock()
}

func (g *Group) ID() string {
	g.RLock()
	defer g.RUnlock()
	return g.idEncoded
}

func (g *Group) GetLabel() string {
	g.RLock()
	defer g.RUnlock()
	return stripNull(string(g.label[:]))
}

func (g *Group) Devices() (devices []common.Device) {
	g.RLock()
	defer g.RUnlock()
	if len(g.devices) == 0 {
		return devices
	}
	for _, dev := range g.devices {
		devices = append(devices, dev.(common.Device))
	}

	return devices
}

func (g *Group) Lights() []common.Light {
	devices := g.Devices()
	lights := make([]common.Light, 0)
	for _, dev := range devices {
		if light, ok := dev.(common.Light); ok {
			lights = append(lights, light)
		}
	}

	return lights
}

func (g *Group) AddDevice(dev GenericDevice) error {
	g.RLock()
	d, ok := g.devices[dev.ID()]
	g.RUnlock()
	if ok {
		// If the old device is of a more specific type than *Device, we don't
		// need to update it
		if _, oldIsDev := d.(*Device); !oldIsDev {
			return common.ErrDuplicate
		}
	}

	g.Lock()
	g.devices[dev.ID()] = dev
	g.Unlock()
	if err := g.addDeviceSubscription(dev); err != nil {
		return err
	}

	g.Notify(common.EventNewDevice{Device: dev})

	return nil
}

func (g *Group) addDeviceSubscription(dev GenericDevice) error {
	sub := dev.Subscribe()
	events := sub.Events()

	go func() {
		for {
			select {
			case <-g.quitChan:
				return
			default:
			}
			select {
			case <-g.quitChan:
				return
			case event := <-events:
				switch event.(type) {
				case common.EventUpdateColor:
					// trigger event if necessary
					_ = g.CachedColor()
				case common.EventUpdatePower:
					// trigger event if necessary
					_ = g.CachedPower()
				}
			}
		}
	}()

	return nil
}

func (g *Group) RemoveDevice(dev GenericDevice) error {
	g.RLock()
	_, ok := g.devices[dev.ID()]
	g.RUnlock()
	if !ok {
		return common.ErrNotFound
	}

	g.Lock()
	delete(g.devices, dev.ID())
	g.Unlock()

	g.Notify(common.EventExpiredDevice{Device: dev})
	return nil
}

func (g *Group) GetPower() (bool, error) {
	return g.getPower(false)
}

func (g *Group) CachedPower() bool {
	p, _ := g.getPower(true)
	return p
}

func (g *Group) getPower(cached bool) (bool, error) {
	var state uint
	g.RLock()
	lastPower := g.power
	g.RUnlock()

	devices := g.Devices()

	if len(devices) == 0 {
		return false, nil
	}

	for _, dev := range devices {
		var (
			p   bool
			err error
		)
		if cached {
			p = dev.CachedPower()
		} else {
			p, err = dev.GetPower()
			if err != nil {
				return false, err
			}
		}
		if p {
			state++
		}
	}

	g.Lock()
	g.power = state > 0
	g.Unlock()
	g.RLock()
	defer g.RUnlock()
	if lastPower != g.power {
		g.Notify(common.EventUpdatePower{Power: g.power})
	}

	return g.power, nil
}

func (g *Group) GetColor() (common.Color, error) {
	return g.getColor(false)
}

func (g *Group) CachedColor() common.Color {
	c, _ := g.getColor(true)
	return c
}

// getColor returns the average color for lights in the group, or error if any
// light returns an error.
func (g *Group) getColor(cached bool) (common.Color, error) {
	var err error

	g.RLock()
	lastColor := g.color
	g.RUnlock()

	lights := g.Lights()

	if len(lights) == 0 {
		return lastColor, nil
	}

	colors := make([]common.Color, len(lights))

	for i, light := range lights {
		var c common.Color
		if cached {
			c = light.CachedColor()
		} else {
			c, err = light.GetColor()
			if err != nil {
				return lastColor, err
			}
		}
		colors[i] = c
	}

	g.Lock()
	g.color = common.AverageColor(colors...)
	g.Unlock()
	g.RLock()
	defer g.RUnlock()
	if !common.ColorEqual(lastColor, g.color) {
		g.Notify(common.EventUpdateColor{Color: g.color})
	}

	return g.color, nil
}

func (g *Group) SetColor(color common.Color, duration time.Duration) error {
	var (
		wg       sync.WaitGroup
		err      error
		errMutex sync.Mutex
	)

	lights := g.Lights()

	if len(lights) == 0 {
		return nil
	}

	for _, light := range lights {
		wg.Add(1)
		go func(light common.Light) {
			e := light.SetColor(color, duration)
			errMutex.Lock()
			if err == nil && e != nil {
				err = e
			}
			errMutex.Unlock()
			wg.Done()
		}(light)
	}

	wg.Wait()
	return err
}

func (g *Group) SetPower(state bool) error {
	var (
		wg       sync.WaitGroup
		err      error
		errMutex sync.Mutex
	)

	devices := g.Devices()

	if len(devices) == 0 {
		return nil
	}

	for _, device := range devices {
		wg.Add(1)
		go func(device common.Device) {
			e := device.SetPower(state)
			errMutex.Lock()
			if err == nil && e != nil {
				err = e
			}
			errMutex.Unlock()
			wg.Done()
		}(device)
	}

	wg.Wait()
	return err
}

func (g *Group) SetPowerDuration(state bool, duration time.Duration) error {
	var (
		wg       sync.WaitGroup
		err      error
		errMutex sync.Mutex
	)

	lights := g.Lights()

	if len(lights) == 0 {
		return nil
	}

	for _, light := range lights {
		wg.Add(1)
		go func(light common.Light) {
			e := light.SetPowerDuration(state, duration)
			errMutex.Lock()
			if err == nil && e != nil {
				err = e
			}
			errMutex.Unlock()
			wg.Done()
		}(light)
	}

	wg.Wait()
	return err
}

func (g *Group) Parse(pkt *packet.Packet) error {
	var shouldUpdate, labelUpdate bool

	s := stateGroup{}
	if err := pkt.DecodePayload(&s); err != nil {
		return err
	}

	g.RLock()
	if s.UpdatedAt > g.updatedAt {
		shouldUpdate = true
	}
	g.RUnlock()

	if shouldUpdate {
		g.Lock()
		g.id = s.ID
		g.idEncoded = strings.Replace(
			base64.URLEncoding.EncodeToString(s.ID[:]),
			`=`, ``, -1,
		)
		g.updatedAt = s.UpdatedAt
		if g.label != s.Label {
			g.label = s.Label
			labelUpdate = true
		}
		g.Unlock()

		if labelUpdate {
			g.Notify(common.EventUpdateLabel{Label: g.GetLabel()})
		}
	}

	return nil
}

// Close cleans up Group resources
func (g *Group) Close() error {
	g.Lock()
	defer g.Unlock()

	select {
	case <-g.quitChan:
		common.Log.Warnf(`group already closed`)
		return common.ErrClosed
	default:
		close(g.quitChan)
	}

	return g.SubscriptionProvider.Close()
}

func NewGroup(pkt *packet.Packet) (*Group, error) {
	g := new(Group)
	g.init()
	err := g.Parse(pkt)

	return g, err
}
