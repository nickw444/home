package protocol

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/nickw444/miio-go/common"
	"github.com/nickw444/miio-go/device"
	"github.com/nickw444/miio-go/protocol/packet"
	"github.com/nickw444/miio-go/protocol/transport"
	"github.com/nickw444/miio-go/subscription"
)

type Protocol interface {
	subscription.SubscriptionTarget

	Discover() error
	SetExpiryTime(duration time.Duration)
}

type protocol struct {
	subscription.SubscriptionTarget
	port          int
	expireAfter   time.Duration
	clock         clock.Clock
	lastDiscovery time.Time

	broadcastDev device.Device
	quitChan     chan struct{}
	devicesMutex sync.RWMutex
	devices      map[uint32]device.Device

	transport     transport.Transport
	deviceFactory DeviceFactory
	cryptoFactory CryptoFactory
}

type DeviceFactory func(deviceId uint32, outbound transport.Outbound, seen time.Time) device.Device
type CryptoFactory func(deviceID uint32, deviceToken []byte, initialStamp uint32, stampTime time.Time) (packet.Crypto, error)

type ProtocolConfig struct {
	ListenPort int
}

func NewProtocol(c *ProtocolConfig) (Protocol, error) {
	clk := clock.New()
	var listenAddr *net.UDPAddr
	if c != nil && c.ListenPort != 0 {
		listenAddr = &net.UDPAddr{Port: c.ListenPort}
	}

	s, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	t := transport.NewTransport(s)
	deviceFactory := func(deviceId uint32, outbound transport.Outbound, seen time.Time) device.Device {
		return device.New(deviceId, outbound, seen)
	}
	cryptoFactory := func(deviceID uint32, deviceToken []byte, initialStamp uint32, stampTime time.Time) (packet.Crypto, error) {
		return packet.NewCrypto(deviceID, deviceToken, initialStamp, stampTime, clk)
	}

	p := newProtocol(clk, t, deviceFactory, cryptoFactory, subscription.NewTarget())
	p.start()
	return p, nil
}

func newProtocol(c clock.Clock, transport transport.Transport, deviceFactory DeviceFactory,
	crptoFactory CryptoFactory, target subscription.SubscriptionTarget) *protocol {

	p := &protocol{
		SubscriptionTarget: target,
		transport:          transport,
		deviceFactory:      deviceFactory,
		cryptoFactory:      crptoFactory,
		clock:              c,
		quitChan:           make(chan struct{}),
		devices:            make(map[uint32]device.Device),
	}

	p.broadcastDev = p.makeBroadcastDev()
	return p
}

func (p *protocol) start() {
	go p.dispatcher()
}

func (p *protocol) makeBroadcastDev() device.Device {
	addr := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: 54321,
	}
	t := p.transport.NewOutbound(nil, addr)
	return p.deviceFactory(0, t, time.Time{})
}

func (p *protocol) SetExpiryTime(duration time.Duration) {
	p.expireAfter = duration
}

func (p *protocol) dispatcher() {
	pkts := p.transport.Inbound().Packets()
	for {
		select {
		case <-p.quitChan:
			return
		default:
		}

		select {
		case <-p.quitChan:
			return
		case pkt := <-pkts:
			go p.process(pkt)
		}
	}
}

func (p *protocol) Discover() error {
	common.Log.Debug("Running discovery...")

	if p.lastDiscovery.After(time.Time{}) {
		// If the device has not been seen recently, it should be expired.
		cutoff := time.Now().Add(p.expireAfter * -1)
		var expiredDevices []device.Device
		p.devicesMutex.RLock()
		for _, dev := range p.devices {
			if dev.Seen().Before(cutoff) {
				common.Log.Infof("Last Seen: %s, Cutoff: %s", dev.Seen(), cutoff)
				expiredDevices = append(expiredDevices, dev)
			}
		}
		p.devicesMutex.RUnlock()

		for _, dev := range expiredDevices {
			common.Log.Debugf("Removing expired device with id %d.", dev.ID())
			p.removeDevice(dev.ID())
			dev.Close()
			err := p.Publish(common.EventExpiredDevice{dev})
			if err != nil {
				common.Log.Warn(err)
			}
		}
	}
	if err := p.broadcastDev.Discover(); err != nil {
		return err
	}

	p.lastDiscovery = time.Now()
	return nil
}
func (p *protocol) process(pkt *packet.Packet) {
	common.Log.Debugf("Processing incoming packet from %s", pkt.Meta.Addr)

	dev := p.getDevice(pkt.Header.DeviceID)
	if dev == nil && pkt.DataLength() == 0 {

		// Device response to a Hello packet.
		crypto, err := p.cryptoFactory(pkt.Header.DeviceID, pkt.Header.Checksum, pkt.Header.Stamp,
			pkt.Meta.DecodeTime)
		if err != nil {
			panic(err)
		}

		t := p.transport.NewOutbound(crypto, pkt.Meta.Addr)
		baseDev := p.deviceFactory(pkt.Header.DeviceID, t, pkt.Meta.DecodeTime)

		// Store the provisional device for now to ensure it can handle subsequent
		// packets that may occur during classification.
		p.addDevice(baseDev)

		dev, err := device.Classify(baseDev)
		if err != nil {
			panic(err)
		}

		// Store the specific device and publish a new device event.
		p.addDevice(dev)
		p.Publish(common.EventNewDevice{Device: dev})
	} else if dev != nil {
		err := dev.Handle(pkt)
		if err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("Unable to process packet."))
	}
}

func (p *protocol) removeDevice(id uint32) {
	p.devicesMutex.Lock()
	delete(p.devices, id)
	p.devicesMutex.Unlock()
}

func (p *protocol) addDevice(dev device.Device) {
	p.devicesMutex.Lock()
	p.devices[dev.ID()] = dev
	p.devicesMutex.Unlock()
}

func (p *protocol) getDevice(id uint32) device.Device {
	p.devicesMutex.RLock()
	dev, ok := p.devices[id]
	p.devicesMutex.RUnlock()
	if !ok {
		return nil
	}
	return dev
}
