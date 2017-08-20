package protocol

import (
	"net"
	"sync"
	"time"

	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol/v2/device"
	"github.com/pdf/golifx/protocol/v2/packet"
	"github.com/pdf/golifx/protocol/v2/shared"
)

// V2 implements the LIFX LAN protocol version 2.
type V2 struct {
	// Port determines UDP port for this protocol instance
	Port int
	// Reliable enables reliable comms, requests ACKs for all operations to
	// ensure they're delivered (recommended)
	Reliable      bool
	initialized   bool
	socket        *net.UDPConn
	timeout       *time.Duration
	retryInterval *time.Duration
	broadcast     *device.Light
	lastDiscovery time.Time
	deviceQueue   chan device.GenericDevice
	wg            sync.WaitGroup
	devices       map[uint64]device.GenericDevice
	subscriptions map[string]*common.Subscription
	locations     map[string]*device.Location
	groups        map[string]*device.Group
	quitChan      chan struct{}
	sync.RWMutex
}

// NewSubscription returns a new *common.Subscription for receiving events from
// this protocol.
func (p *V2) NewSubscription() (*common.Subscription, error) {
	if err := p.init(); err != nil {
		return nil, err
	}
	sub := common.NewSubscription(p)
	p.Lock()
	p.subscriptions[sub.ID()] = sub
	p.Unlock()
	return sub, nil
}

// CloseSubscription is a callback for handling the closing of subscriptions.
func (p *V2) CloseSubscription(sub *common.Subscription) error {
	p.RLock()
	_, ok := p.subscriptions[sub.ID()]
	p.RUnlock()
	if !ok {
		return common.ErrNotFound
	}
	p.Lock()
	delete(p.subscriptions, sub.ID())
	p.Unlock()

	return nil
}

func (p *V2) init() error {
	p.RLock()
	if p.initialized {
		p.RUnlock()
		return nil
	}
	p.RUnlock()

	p.Lock()
	defer p.Unlock()
	if p.Port == 0 {
		p.Port = shared.DefaultPort
	}
	socket, err := net.ListenUDP(`udp4`, &net.UDPAddr{Port: p.Port})
	if err != nil {
		return err
	}
	p.socket = socket
	addr := net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: shared.DefaultPort,
	}
	broadcastDev, err := device.New(&addr, p.socket, p.timeout, p.retryInterval, false, nil)
	if err != nil {
		return err
	}
	p.broadcast = &device.Light{Device: broadcastDev}
	broadcastSub, err := p.broadcast.NewSubscription()
	if err != nil {
		return err
	}
	p.deviceQueue = make(chan device.GenericDevice, 16)
	p.devices = make(map[uint64]device.GenericDevice)
	p.locations = make(map[string]*device.Location)
	p.groups = make(map[string]*device.Group)
	p.subscriptions = make(map[string]*common.Subscription)
	p.quitChan = make(chan struct{})
	go p.broadcastLimiter(broadcastSub.Events())
	go p.dispatcher()
	go p.addDevices()
	p.initialized = true

	return nil
}

// Pushes an event to subscribers
func (p *V2) publish(event interface{}) error {
	p.RLock()
	subs := make(map[string]*common.Subscription, len(p.subscriptions))
	for k, sub := range p.subscriptions {
		subs[k] = sub
	}
	p.RUnlock()

	for _, sub := range subs {
		if err := sub.Write(event); err != nil {
			return err
		}
	}

	return nil
}

// SetTimeout attaches a timeout to the protocol
func (p *V2) SetTimeout(timeout *time.Duration) {
	p.Lock()
	p.timeout = timeout
	p.Unlock()
}

// SetRetryInterval attaches a retry interval to the protocol
func (p *V2) SetRetryInterval(retryInterval *time.Duration) {
	p.Lock()
	p.retryInterval = retryInterval
	p.Unlock()
}

// Discover initiates device discovery, this may be a noop in some future
// protocol versions.  This is called immediately when the client connects to
// the protocol
func (p *V2) Discover() error {
	if err := p.init(); err != nil {
		return err
	}
	if p.lastDiscovery.After(time.Time{}) {
		var extinct []device.GenericDevice
		p.RLock()
		for _, dev := range p.devices {
			// If the device has not been seen in twice the time since the last
			// discovery, mark it as extinct
			if dev.Seen().Before(time.Now().Add(time.Since(p.lastDiscovery) * -2)) {
				extinct = append(extinct, dev)
			}
		}
		p.RUnlock()
		// Remove extinct devices
		for _, dev := range extinct {
			p.removeDevice(dev.ID())

			locationID := dev.CachedLocation()
			location, err := p.getLocation(locationID)
			if err == nil {
				if err = location.RemoveDevice(dev); err != nil {
					common.Log.Warnf("Failed removing extinct device '%d' from location (%s): %v", dev.ID(), locationID, err)
				}
				if len(location.Devices()) == 0 {
					p.removeLocation(location.ID())
					if err = p.publish(common.EventExpiredLocation{Location: location}); err != nil {
						common.Log.Warnf("Failed publishing expired event for location '%s'", locationID)
					}
				}
			}

			groupID := dev.CachedGroup()
			group, err := p.getGroup(groupID)
			if err == nil {
				if err = group.RemoveDevice(dev); err != nil {
					common.Log.Warnf("Failed removing extinct device '%d' from group (%s): %v", dev.ID(), groupID, err)
				}
				if len(group.Devices()) == 0 {
					p.removeGroup(group.ID())
					if err = p.publish(common.EventExpiredGroup{Group: group}); err != nil {
						common.Log.Warnf("Failed publishing expired event for group '%s'", groupID)
					}
				}
			}

			err = p.publish(common.EventExpiredDevice{Device: dev})
			if err != nil {
				common.Log.Warnf("Failed removing extinct device '%d' from client: %v", dev.ID(), err)
			}
		}
	}
	if err := p.broadcast.Discover(); err != nil {
		return err
	}
	p.Lock()
	p.lastDiscovery = time.Now()
	p.Unlock()

	return nil
}

// SetPower sets the power state globally, on all devices
func (p *V2) SetPower(state bool) error {
	p.RLock()
	defer p.RUnlock()
	for _, dev := range p.devices {
		if err := dev.SetPower(state); err != nil {
			common.Log.Warnf("Failed setting power on %d: %+v", dev.ID(), err)
			continue
		}
	}
	return nil
}

// SetPowerDuration sets the power state globally, on all devices, transitioning
// over the specified duration
func (p *V2) SetPowerDuration(state bool, duration time.Duration) error {
	p.RLock()
	defer p.RUnlock()
	for _, dev := range p.devices {
		l, ok := dev.(*device.Light)
		if !ok {
			continue
		}
		if err := l.SetPowerDuration(state, duration); err != nil {
			common.Log.Warnf("Failed setting power on %d: %+v", l.ID(), err)
			continue
		}
	}
	return nil
}

// SetColor changes the color globally, on all lights, transitioning over the
// specified duration
func (p *V2) SetColor(color common.Color, duration time.Duration) error {
	p.RLock()
	defer p.RUnlock()
	for _, dev := range p.devices {
		l, ok := dev.(*device.Light)
		if !ok {
			continue
		}
		if err := l.SetColor(color, duration); err != nil {
			common.Log.Warnf("Failed setting color on %d: %+v", l.ID(), err)
			continue
		}
	}
	return nil
}

// Close closes the protocol driver, no further communication with the protocol
// is possible
func (p *V2) Close() error {
	for _, sub := range p.subscriptions {
		if err := sub.Close(); err != nil {
			return err
		}
	}

	p.Lock()
	defer p.Unlock()

	for _, location := range p.locations {
		if err := location.Close(); err != nil {
			return err
		}
	}

	for _, group := range p.groups {
		if err := group.Close(); err != nil {
			return err
		}
	}

	for _, dev := range p.devices {
		if err := dev.Close(); err != nil {
			return err
		}
	}

	if err := p.broadcast.Close(); err != nil {
		return err
	}

	select {
	case <-p.quitChan:
		common.Log.Warnf(`protocol already closed`)
		return common.ErrClosed
	default:
		close(p.quitChan)
		p.wg.Wait()
		close(p.deviceQueue)
	}

	return nil
}

func (p *V2) broadcastLimiter(events <-chan interface{}) {
	for {
		select {
		case <-p.quitChan:
			return
		default:
		}
		select {
		case <-p.quitChan:
			return
		case event := <-events:
			switch event.(type) {
			case shared.EventBroadcastSent:
				p.RLock()
				for _, dev := range p.devices {
					dev.ResetLimiter()
				}
				p.RUnlock()
			case shared.EventRequestSent:
				p.broadcast.ResetLimiter()
			}
		}
	}
}

func (p *V2) dispatcher() {
	for {
		select {
		case <-p.quitChan:
			p.Lock()
			for _, dev := range p.devices {
				if err := dev.Close(); err != nil {
					common.Log.Errorf("Failed closing device '%d': %v", dev.ID(), err)
				}
			}
			if err := p.socket.Close(); err != nil {
				common.Log.Errorf("Failed closing socket: %v", err)
			}
			p.Unlock()
			return
		default:
			buf := make([]byte, 1500)
			n, addr, err := p.socket.ReadFromUDP(buf)
			if err != nil {
				common.Log.Errorf("Failed reading from socket: %v", err)
				continue
			}
			pkt, err := packet.Decode(buf[:n])
			if err != nil {
				common.Log.Errorf("Failed decoding packet: %v", err)
				continue
			}
			go p.process(pkt, addr)
		}
	}
}

// GetLocations returns a slice of all locations known to the protocol, or
// common.ErrNotFound if no locations are currently known.
func (p *V2) GetLocations() ([]common.Location, error) {
	p.RLock()
	defer p.RUnlock()
	if len(p.locations) == 0 {
		return nil, common.ErrNotFound
	}
	locations := make([]common.Location, len(p.locations))
	i := 0
	for _, location := range p.locations {
		locations[i] = location
		i++
	}

	return locations, nil
}

// GetGroups returns a slice of all groups known to the client, or
// common.ErrNotFound if no groups are currently known.
func (p *V2) GetGroups() ([]common.Group, error) {
	p.RLock()
	defer p.RUnlock()
	if len(p.groups) == 0 {
		return nil, common.ErrNotFound
	}
	groups := make([]common.Group, len(p.groups))
	i := 0
	for _, group := range p.groups {
		groups[i] = group
		i++
	}

	return groups, nil
}

// GetDevices returns a slice of all devices known to the protocol, or
// common.ErrNotFound if no devices are currently known.
func (p *V2) GetDevices() ([]common.Device, error) {
	p.RLock()
	defer p.RUnlock()
	if len(p.devices) == 0 {
		return nil, common.ErrNotFound
	}
	devices := make([]common.Device, len(p.devices))
	i := 0
	for _, device := range p.devices {
		devices[i] = device
		i++
	}

	return devices, nil
}

func (p *V2) GetLocation(id string) (common.Location, error) {
	return p.getLocation(id)
}

func (p *V2) getLocation(id string) (*device.Location, error) {
	p.RLock()
	location, ok := p.locations[id]
	p.RUnlock()
	if !ok {
		return nil, common.ErrNotFound
	}

	return location, nil
}

func (p *V2) GetGroup(id string) (common.Group, error) {
	return p.getGroup(id)
}

func (p *V2) getGroup(id string) (*device.Group, error) {
	p.RLock()
	group, ok := p.groups[id]
	p.RUnlock()
	if !ok {
		return nil, common.ErrNotFound
	}

	return group, nil
}

func (p *V2) GetDevice(id uint64) (common.Device, error) {
	return p.getDevice(id)
}

func (p *V2) getDevice(id uint64) (device.GenericDevice, error) {
	p.RLock()
	dev, ok := p.devices[id]
	p.RUnlock()
	if !ok {
		return nil, common.ErrNotFound
	}

	return dev, nil
}

func (p *V2) process(pkt *packet.Packet, addr *net.UDPAddr) {
	common.Log.Debugf("Processing packet from %v: source %d, type %d, sequence %d, target %d, tagged %v, resRequired %v, ackRequired %v", addr.IP, pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())

	// Update device seen time for any targeted packets
	if pkt.Target != 0 {
		dev, err := p.getDevice(pkt.Target)
		if err == nil {
			dev.SetSeen(time.Now())
		}
	}

	// Broadcast packets, or packets generated by other clients
	if pkt.GetSource() != packet.ClientID {
		switch pkt.GetType() {
		case device.StatePower:
			dev, err := p.getDevice(pkt.GetTarget())
			if err != nil {
				common.Log.Debugf("Skipping StatePower packet for unknown device: source %d, type %d, sequence %d, target %d, tagged %v, resRequired %v, ackRequired %v", pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())
				return
			}
			err = dev.SetStatePower(pkt)
			if err != nil {
				common.Log.Debugf("Failed setting StatePower on device: source %d, type %d, sequence %d, target %d, tagged %v, resRequired %v, ackRequired %v", pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())
				return
			}
		case device.StateLabel:
			dev, err := p.getDevice(pkt.GetTarget())
			if err != nil {
				common.Log.Debugf("Skipping StateLabel packet for unknown device: source %d, type %d, sequence %d, target %d, tagged %v, resRequired %v, ackRequired %v", pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())
				return
			}
			err = dev.SetStateLabel(pkt)
			if err != nil {
				common.Log.Debugf("Failed setting StatePower on device: source %v, type %v, sequence %v, target %v, tagged %v, resRequired %v, ackRequired %v", pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())
				return
			}
		case device.State:
			dev, err := p.getDevice(pkt.GetTarget())
			if err != nil {
				common.Log.Debugf("Skipping State packet for unknown device: source %d, type %d, sequence %d, target %d, tagged %v, resRequired %v, ackRequired %v", pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())
				return
			}
			light, ok := dev.(*device.Light)
			if !ok {
				common.Log.Debugf("Skipping State packet for non-light device: source %d, type %d, sequence %d, target %d, tagged %v, resRequired %v, ackRequired %v", pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())
				return
			}
			err = light.SetState(pkt)
			if err != nil {
				common.Log.Debugf("Failed setting State on device: source %v, type %v, sequence %v, target %v, tagged %v, resRequired %v, ackRequired %v", pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())
				return
			}
		default:
			common.Log.Debugf("Skipping packet with non-local source: source %v, type %v, sequence %v, target %v, tagged %v, resRequired %v, ackRequired %v", pkt.GetSource(), pkt.GetType(), pkt.GetSequence(), pkt.GetTarget(), pkt.GetTagged(), pkt.GetResRequired(), pkt.GetAckRequired())
		}
		return
	}

	// Packets processed at the protocol level regardless of target
	switch pkt.GetType() {
	case device.StateLocation:
		p.addLocation(pkt)
	case device.StateGroup:
		p.addGroup(pkt)
	}

	// Packets processed at the protocol level or returned to target
	switch pkt.GetType() {
	case device.StateService:
		dev, err := p.getDevice(pkt.Target)
		if err != nil {
			// New device
			dev, err = device.New(addr, p.socket, p.timeout, p.retryInterval, p.Reliable, pkt)
			if err != nil {
				common.Log.Errorf("Failed creating device: %v", err)
				return
			}
		}
		p.wg.Add(1)
		p.deviceQueue <- dev
		p.wg.Done()
	default:
		if pkt.GetTarget() == 0 {
			common.Log.Debugf("Skipping packet without target")
			return
		}
		dev, err := p.getDevice(pkt.GetTarget())
		if err != nil {
			common.Log.Errorf("No known device with ID %d", pkt.GetTarget())
			return
		}
		common.Log.Debugf("Returning packet to device %d", dev.ID())
		dev.Handle(pkt)
	}
}

func (p *V2) addLocation(pkt *packet.Packet) {
	l, err := device.NewLocation(pkt)
	if err != nil {
		common.Log.Errorf("Error parsing location: %v", err)
		return
	}
	location, err := p.getLocation(l.ID())
	if err == nil {
		if err := location.Parse(pkt); err != nil {
			common.Log.Errorf("Error parsing location: %v", err)
		}
		return
	}

	p.Lock()
	p.locations[l.ID()] = l
	p.Unlock()
	if err := p.publish(common.EventNewLocation{Location: l}); err != nil {
		common.Log.Errorf("Error adding location to client: %v", err)
		return
	}
}

func (p *V2) removeLocation(id string) {
	p.Lock()
	delete(p.locations, id)
	p.Unlock()
}

func (p *V2) addGroup(pkt *packet.Packet) {
	g, err := device.NewGroup(pkt)
	if err != nil {
		common.Log.Errorf("Error parsing group: %v", err)
		return
	}
	group, err := p.getGroup(g.ID())
	if err == nil {
		if err := group.Parse(pkt); err != nil {
			common.Log.Errorf("Error parsing group: %v", err)
		}
		return
	}

	p.Lock()
	p.groups[g.ID()] = g
	p.Unlock()
	if err := p.publish(common.EventNewGroup{Group: g}); err != nil {
		common.Log.Errorf("Error adding group to client: %v", err)
		return
	}
}

func (p *V2) removeGroup(id string) {
	p.Lock()
	delete(p.groups, id)
	p.Unlock()
}

func (p *V2) addDevices() {
	for dev := range p.deviceQueue {
		p.addDevice(dev)
		// Perform state discovery on lights
		if l, ok := dev.(*device.Light); ok {
			if err := l.Get(); err != nil {
				common.Log.Debugf("Failed getting light state: %v", err)
			}
		}
	}
}

func (p *V2) addDevice(dev device.GenericDevice) {
	common.Log.Debugf("Attempting to add device: %d", dev.ID())
	d, err := p.getDevice(dev.ID())
	known := err == nil
	if known {
		dev = d
	} else {
		// We don't know this device, add it now and possibly overwrite it
		// later, this is necessary so that we can deliver responses to queries
		// for device type information
		p.Lock()
		p.devices[dev.ID()] = dev
		p.Unlock()
	}

	if dev.Provisional() {
		// Determine device type
		dev = p.classifyDevice(dev)
	}

	p.updateLocationGroup(dev)

	if known {
		common.Log.Debugf("Device already known: %d", dev.ID())
		return
	}

	sub, err := dev.NewSubscription()
	if err != nil {
		common.Log.Warnf("Error obtaining subscription from %d", dev.ID())
	} else {
		go p.broadcastLimiter(sub.Events())
	}

	common.Log.Debugf("Adding device to client: %d", dev.ID())
	if err := p.publish(common.EventNewDevice{Device: dev}); err != nil {
		common.Log.Errorf("Error adding device to client: %v", err)
		return
	}
	common.Log.Debugf("Added device to client: %d", dev.ID())
}

func (p *V2) removeDevice(id uint64) {
	p.Lock()
	delete(p.devices, id)
	p.Unlock()
}

func (p *V2) updateLocationGroup(dev device.GenericDevice) {
	groupID, err := dev.GetGroup()
	if err != nil {
		common.Log.Warnf("Error retrieving device group: %v", err)
		return
	}
	group, err := p.getGroup(groupID)
	if err != nil {
		common.Log.Warnf("Unknown group ID: %s", groupID)
		return
	}
	common.Log.Debugf("Adding device to group (%s): %v", groupID, dev.ID())
	if err = group.AddDevice(dev); err != nil {
		common.Log.Debugf("Error adding device to group: %v", err)
	}

	locationID, err := dev.GetLocation()
	if err != nil {
		common.Log.Warnf("Error retrieving device location: %v", err)
		return
	}
	p.RLock()
	location, ok := p.locations[locationID]
	p.RUnlock()
	if !ok {
		common.Log.Warnf("Unknown location ID: %v", locationID)
		return
	}
	common.Log.Debugf("Adding device to location (%s): %v", locationID, dev.ID())
	if err = location.AddDevice(dev); err != nil {
		common.Log.Debugf("Error adding device to location: %v", err)
	}
}

// classifyDevice either constructs a device.Light from the passed dev, or returns
// the dev untouched
func (p *V2) classifyDevice(dev device.GenericDevice) device.GenericDevice {
	common.Log.Debugf("Attempting to determine device type for: %d", dev.ID())
	vendor, err := dev.GetHardwareVendor()
	if err != nil {
		common.Log.Errorf("Error retrieving device hardware vendor: %v", err)
		return dev
	}
	product, err := dev.GetHardwareProduct()
	if err != nil {
		common.Log.Errorf("Error retrieving device hardware product: %v", err)
		return dev
	}

	defer dev.SetProvisional(false)

	switch vendor {
	case device.VendorLifx:
		switch product {
		case device.ProductLifxOriginal1000, device.ProductLifxColor650, device.ProductLifxWhite800LowVoltage, device.ProductLifxWhite800HighVoltage, device.ProductLifxWhite900BR30, device.ProductLifxColor1000BR30, device.ProductLifxColor1000:
			p.Lock()
			d := dev.(*device.Device)
			d.Lock()
			l := &device.Light{Device: d}
			common.Log.Debugf("Device is a light: %v", l.ID())
			// Replace the known dev with our constructed light
			p.devices[l.ID()] = l
			d.Unlock()
			p.Unlock()
			return l
		}
	}

	return dev
}
