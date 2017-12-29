// Package device implements a LIFX LAN protocol version 2 device.
//
// This package is not designed to be accessed by end users, all interaction
// should occur via the Client in the golifx package.
package device

import (
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol/v2/packet"
	"github.com/pdf/golifx/protocol/v2/shared"
)

const (
	GetService        shared.Message = 2
	StateService      shared.Message = 3
	GetHostInfo       shared.Message = 12
	StateHostInfo     shared.Message = 13
	GetHostFirmware   shared.Message = 14
	StateHostFirmware shared.Message = 15
	GetWifiInfo       shared.Message = 16
	StateWifiInfo     shared.Message = 17
	GetWifiFirmware   shared.Message = 18
	StateWifiFirmware shared.Message = 19
	GetPower          shared.Message = 20
	SetPower          shared.Message = 21
	StatePower        shared.Message = 22
	GetLabel          shared.Message = 23
	SetLabel          shared.Message = 24
	StateLabel        shared.Message = 25
	GetVersion        shared.Message = 32
	StateVersion      shared.Message = 33
	GetInfo           shared.Message = 34
	StateInfo         shared.Message = 35
	Acknowledgement   shared.Message = 45
	GetLocation       shared.Message = 48
	StateLocation     shared.Message = 50
	GetGroup          shared.Message = 51
	StateGroup        shared.Message = 53
	EchoRequest       shared.Message = 58
	EchoResponse      shared.Message = 59
)

type response struct {
	ch   packet.Chan
	done doneChan
	wg   sync.WaitGroup
}

type doneChan chan struct{}
type responseMap map[uint8]*response

type Device struct {
	id                    uint64
	address               *net.UDPAddr
	power                 uint16
	label                 string
	hardwareVersion       stateVersion
	firmwareVersion       uint32
	firmwareVersionString string
	provisional           bool
	product               *Product

	locationID string
	groupID    string

	sequence      uint8
	requestSocket *net.UDPConn
	responseMap   responseMap
	responseInput packet.Chan
	quitChan      chan struct{}
	timeout       *time.Duration
	retryInterval *time.Duration
	limiter       *time.Timer
	seen          time.Time
	reliable      bool
	common.SubscriptionProvider
	sync.RWMutex
}

type stateService struct {
	Service shared.Service `struc:"little"`
	Port    uint32         `struc:"little"`
}

type stateVersion struct {
	Vendor  uint32 `struc:"little"`
	Product uint32 `struc:"little"`
	Version uint32 `struc:"little"`
}

type stateLabel struct {
	Label [32]byte `struc:"little"`
}

type statePower struct {
	Level uint16 `struc:"little"`
}

type stateHostFirmware struct {
	Build    uint64 `struc:"little"`
	Reserved uint64 `struc:"little"`
	Version  uint32 `struc:"little"`
}

type payloadPower struct {
	Level uint16 `struc:"little"`
}

type payloadLabel struct {
	Label [32]byte `struc:"little"`
}

func (f *stateHostFirmware) String() string {
	return fmt.Sprintf("%d.%d", (f.Version&0xffff0000)>>16, f.Version&0xffff)
}

func (d *Device) init(addr *net.UDPAddr, requestSocket *net.UDPConn, timeout *time.Duration, retryInterval *time.Duration, reliable bool) {
	d.Lock()
	d.address = addr
	d.requestSocket = requestSocket
	d.timeout = timeout
	d.retryInterval = retryInterval
	d.reliable = reliable
	d.limiter = time.NewTimer(shared.RateLimit)
	d.responseMap = make(responseMap)
	d.responseInput = make(packet.Chan, 32)
	d.quitChan = make(chan struct{})
	d.provisional = true
	d.Unlock()
}

func (d *Device) ID() uint64 {
	return d.id
}

func (d *Device) Discover() error {
	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(GetService)
	_, err := d.Send(pkt, false, false)
	return err
}

func (d *Device) Provisional() bool {
	d.RLock()
	defer d.RUnlock()
	return d.provisional
}

func (d *Device) SetProvisional(provisional bool) {
	d.Lock()
	d.provisional = provisional
	d.Unlock()
}

func (d *Device) SetStateLabel(pkt *packet.Packet) error {
	l := stateLabel{}
	if err := pkt.DecodePayload(&l); err != nil {
		return err
	}
	common.Log.Debugf("Got label (%d): %v", d.id, string(l.Label[:]))
	newLabel := stripNull(string(l.Label[:]))
	if newLabel != d.CachedLabel() {
		d.Lock()
		d.label = newLabel
		d.Unlock()
		d.Notify(common.EventUpdateLabel{Label: newLabel})
	}

	return nil
}

func (d *Device) SetStateLocation(pkt *packet.Packet) error {
	l := &Location{}
	if err := l.Parse(pkt); err != nil {
		return err
	}
	common.Log.Debugf("Got location (%d): %s (%s)", d.id, l.ID(), l.GetLabel())
	newLocation := l.ID()
	if newLocation != d.CachedLocation() {
		d.Lock()
		d.locationID = newLocation
		d.Unlock()
		// TODO: Work out what to notify on without causing protocol version
		// dependency
	}

	return nil
}

func (d *Device) SetStateGroup(pkt *packet.Packet) error {
	g := &Group{}
	if err := g.Parse(pkt); err != nil {
		return err
	}
	common.Log.Debugf("Got group (%d): %s (%s)", d.id, g.ID(), g.GetLabel())
	newGroup := g.ID()
	if newGroup != d.CachedGroup() {
		d.Lock()
		d.groupID = newGroup
		d.Unlock()
		// TODO: Work out what to notify on without causing protocol version
		// dependency
	}

	return nil
}

func (d *Device) SetStateHostFirmware(pkt *packet.Packet) error {
	f := stateHostFirmware{}
	if err := pkt.DecodePayload(&f); err != nil {
		return err
	}
	common.Log.Debugf("Got firmware version (%d): %d", d.id, f.Version)
	d.RLock()
	version := d.firmwareVersion
	d.RUnlock()
	if f.Version != version {
		d.Lock()
		d.firmwareVersion = f.Version
		d.firmwareVersionString = f.String()
		d.Unlock()
	}

	return nil
}

func (d *Device) GetLabel() (string, error) {
	label := d.CachedLabel()
	if len(label) != 0 {
		return label, nil
	}

	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(GetLabel)
	req, err := d.Send(pkt, d.reliable, true)
	if err != nil {
		return ``, err
	}

	common.Log.Debugf("Waiting for label (%d)", d.id)
	pktResponse := <-req
	if pktResponse == nil {
		return ``, common.ErrProtocol
	}
	if pktResponse.Error != nil {
		return ``, err
	}

	err = d.SetStateLabel(pktResponse.Result)
	if err != nil {
		return ``, err
	}

	return d.CachedLabel(), nil
}

func (d *Device) SetLabel(label string) error {
	if d.CachedLabel() == label {
		return nil
	}

	p := &payloadLabel{}
	copy(p.Label[:], label)

	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(SetLabel)
	if err := pkt.SetPayload(p); err != nil {
		return err
	}

	common.Log.Debugf("Setting label on %d: %s", d.id, label)
	req, err := d.Send(pkt, d.reliable, false)
	if err != nil {
		return err
	}
	if d.reliable {
		// Wait for ack
		<-req
		common.Log.Debugf("Setting label on %d acknowledged", d.id)
	}

	d.Lock()
	d.label = label
	d.Unlock()
	d.Notify(common.EventUpdateLabel{Label: label})
	return nil
}

func (d *Device) CachedLabel() string {
	d.RLock()
	defer d.RUnlock()
	return d.label
}

func (d *Device) SetStatePower(pkt *packet.Packet) error {
	p := statePower{}
	if err := pkt.DecodePayload(&p); err != nil {
		return err
	}
	common.Log.Debugf("Got power (%d): %d", d.id, d.power)

	state := p.Level > 0
	if d.CachedPower() != state {
		d.Lock()
		d.power = p.Level
		d.Unlock()
		d.Notify(common.EventUpdatePower{Power: state})
	}

	return nil
}

func (d *Device) GetPower() (bool, error) {
	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(GetPower)
	req, err := d.Send(pkt, d.reliable, true)
	if err != nil {
		return false, err
	}

	common.Log.Debugf("Waiting for power (%d)", d.id)
	pktResponse := <-req
	if pktResponse == nil {
		return false, common.ErrProtocol
	}
	if pktResponse.Error != nil {
		return false, err
	}

	err = d.SetStatePower(pktResponse.Result)
	if err != nil {
		return false, err
	}

	return d.CachedPower(), nil
}

func (d *Device) CachedPower() bool {
	d.RLock()
	defer d.RUnlock()
	return d.power > 0
}

func (d *Device) SetPower(state bool) error {
	p := &payloadPower{}
	if state {
		p.Level = math.MaxUint16
	}

	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(SetPower)
	if err := pkt.SetPayload(p); err != nil {
		return err
	}

	common.Log.Debugf("Setting power state on %d: %v", d.id, state)
	req, err := d.Send(pkt, d.reliable, false)
	if err != nil {
		return err
	}
	if d.reliable {
		// Wait for ack
		<-req
		common.Log.Debugf("Setting power state on %d acknowledged", d.id)
	}

	d.Lock()
	d.power = p.Level
	d.Unlock()
	d.Notify(common.EventUpdatePower{Power: p.Level > 0})
	return nil
}

func (d *Device) CachedLocation() string {
	d.RLock()
	defer d.RUnlock()
	return d.locationID
}

func (d *Device) GetLocation() (ret string, err error) {
	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(GetLocation)
	req, err := d.Send(pkt, d.reliable, true)
	if err != nil {
		return ret, err
	}

	common.Log.Debugf("Waiting for location (%d)", d.id)
	pktResponse := <-req
	if pktResponse == nil {
		return ``, common.ErrProtocol
	}
	if pktResponse.Error != nil {
		return ret, err
	}

	err = d.SetStateLocation(pktResponse.Result)
	if err != nil {
		return ret, err
	}

	return d.CachedLocation(), nil
}

func (d *Device) CachedGroup() string {
	d.RLock()
	defer d.RUnlock()
	return d.groupID
}

func (d *Device) GetGroup() (ret string, err error) {
	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(GetGroup)
	req, err := d.Send(pkt, d.reliable, true)
	if err != nil {
		return ret, err
	}

	common.Log.Debugf("Waiting for group (%d)", d.id)
	pktResponse := <-req
	if pktResponse == nil {
		return ``, common.ErrProtocol
	}
	if pktResponse.Error != nil {
		return ret, err
	}

	err = d.SetStateGroup(pktResponse.Result)
	if err != nil {
		return ret, err
	}

	return d.CachedGroup(), nil
}

func (d *Device) GetHardwareVendor() (uint32, error) {
	if d.CachedHardwareVendor() != 0 {
		return d.CachedHardwareVendor(), nil
	}

	_, err := d.GetHardwareVersion()
	if err != nil {
		return 0, err
	}

	return d.CachedHardwareVendor(), nil
}

func (d *Device) GetHardwareProduct() (uint32, error) {
	if d.CachedHardwareProduct() != 0 {
		return d.CachedHardwareProduct(), nil
	}

	_, err := d.GetHardwareVersion()
	if err != nil {
		return 0, err
	}

	return d.CachedHardwareProduct(), nil
}

func (d *Device) GetHardwareVersion() (uint32, error) {
	if d.CachedHardwareVersion() != 0 {
		return d.CachedHardwareVersion(), nil
	}

	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(GetVersion)
	req, err := d.Send(pkt, d.reliable, true)
	if err != nil {
		return 0, err
	}

	common.Log.Debugf("Waiting for hardware version (%d)", d.id)
	pktResponse := <-req
	if pktResponse == nil {
		return 0, common.ErrProtocol
	}
	if pktResponse.Error != nil {
		return 0, err
	}

	v := stateVersion{}
	if err = pktResponse.Result.DecodePayload(&v); err != nil {
		return 0, err
	}
	common.Log.Debugf("Got hardware version (%d): %+v", d.id, v)

	d.Lock()
	d.hardwareVersion = v
	if vendor, ok := Vendors[v.Vendor]; ok {
		if product, ok := vendor.Products[v.Product]; ok {
			d.product = product
		}
	}
	d.Unlock()

	return d.CachedHardwareVersion(), nil
}

func (d *Device) GetProduct() (*Product, error) {
	if d.CachedProduct() != nil {
		return d.CachedProduct(), nil
	}

	_, err := d.GetHardwareVersion()
	if err != nil {
		return nil, err
	}

	return d.CachedProduct(), nil
}

func (d *Device) CachedHardwareVersion() uint32 {
	d.RLock()
	defer d.RUnlock()
	return d.hardwareVersion.Version
}

func (d *Device) CachedHardwareVendor() uint32 {
	d.RLock()
	defer d.RUnlock()
	return d.hardwareVersion.Vendor
}

func (d *Device) CachedHardwareProduct() uint32 {
	d.RLock()
	defer d.RUnlock()
	return d.hardwareVersion.Product
}

func (d *Device) CachedProduct() *Product {
	d.RLock()
	defer d.RUnlock()
	return d.product
}

func (d *Device) GetProductName() (string, error) {
	p, err := d.GetProduct()
	if err != nil {
		return ``, err
	}
	if p == nil {
		return ``, common.ErrNotFound
	}

	return p.Name, nil
}

func (d *Device) CachedFirmwareVersion() string {
	d.RLock()
	defer d.RUnlock()
	return d.firmwareVersionString
}

func (d *Device) GetFirmwareVersion() (ret string, err error) {
	pkt := packet.New(d.address, d.requestSocket)
	pkt.SetType(GetHostFirmware)
	req, err := d.Send(pkt, d.reliable, true)
	if err != nil {
		return ret, err
	}

	common.Log.Debugf("Waiting for firmware data (%d)", d.id)
	pktResponse := <-req
	if pktResponse == nil {
		return ``, common.ErrProtocol
	}
	if pktResponse.Error != nil {
		return ret, err
	}

	err = d.SetStateHostFirmware(pktResponse.Result)
	if err != nil {
		return ret, err
	}

	return d.CachedFirmwareVersion(), nil
}

func (d *Device) Handle(pkt *packet.Packet) {
	d.responseInput <- &packet.Response{Result: pkt}
}

func (d *Device) GetAddress() *net.UDPAddr {
	return d.address
}

func (d *Device) ResetLimiter() {
	d.Lock()
	d.limiter.Reset(shared.RateLimit)
	d.Unlock()
}

func (d *Device) resetLimiter(broadcast bool) {
	if broadcast {
		d.Notify(shared.EventRequestSent{})
	} else {
		d.Notify(shared.EventBroadcastSent{})
	}
	d.ResetLimiter()
}

func (d *Device) Send(pkt *packet.Packet, ackRequired, responseRequired bool) (packet.Chan, error) {
	proxyChan := make(packet.Chan)

	// Rate limiter
	<-d.limiter.C

	// Broadcast vs direct
	broadcast := d.id == 0
	if broadcast {
		// Broadcast can't be reliable
		ackRequired = false
		pkt.SetTagged(true)
	} else {
		pkt.SetTarget(d.id)
		if ackRequired {
			pkt.SetAckRequired(true)
		}
		if responseRequired {
			pkt.SetResRequired(true)
		}
		if ackRequired || responseRequired {
			seq, res := d.addSeq()
			pkt.SetSequence(seq)

			go func() {
				defer func() {
					select {
					case <-d.quitChan:
					default:
						close(res.done)
						close(proxyChan)
					}
				}()

				var (
					timeout <-chan time.Time
					ticker  = time.NewTicker(*d.retryInterval)
				)

				if d.timeout == nil || *d.timeout == 0 {
					timeout = make(<-chan time.Time)
				} else {
					timeout = time.After(*d.timeout)
				}

				for {
					select {
					case <-d.quitChan:
						return
					default:
					}
					select {
					case <-d.quitChan:
						return
					case pktResponse, ok := <-res.ch:
						if !ok {
							return
						}
						if pktResponse.Result.GetType() == Acknowledgement {
							common.Log.Debugf("Got ACK for seq %d on device %d, cancelling retries", seq, d.ID())
							ticker.Stop()
							// Ack does not resolve outstanding request,
							// continue waiting for response
							if responseRequired {
								continue
							}
						}
						proxyChan <- pktResponse
						return
					case <-ticker.C:
						common.Log.Debugf("Retrying send for seq %d on device %d after %d milliseconds", seq, d.ID(), *d.retryInterval/time.Millisecond)
						if err := pkt.Write(); err != nil {
							proxyChan <- &packet.Response{
								Error: err,
							}
							return
						}
					case <-timeout:
						proxyChan <- &packet.Response{
							Error: common.ErrTimeout,
						}
						return
					}
				}
			}()
		}
	}

	err := pkt.Write()
	d.resetLimiter(broadcast)

	return proxyChan, err
}

func (d *Device) Seen() time.Time {
	d.RLock()
	defer d.RUnlock()
	return d.seen
}

func (d *Device) SetSeen(seen time.Time) {
	d.Lock()
	d.seen = seen
	d.Unlock()
}

// Close cleans up Device resources
func (d *Device) Close() error {
	d.Lock()
	defer d.Unlock()

	select {
	case <-d.quitChan:
		return common.ErrClosed
	default:
		close(d.quitChan)
		for seq, res := range d.responseMap {
			select {
			case res.ch <- &packet.Response{Error: common.ErrClosed}:
			case <-res.done:
			default:
				close(res.done)
			}
			res.wg.Wait()
			close(res.ch)
			delete(d.responseMap, seq)
		}
	}

	return d.SubscriptionProvider.Close()
}

func (d *Device) handler() {
	var ok bool

	for {
		select {
		case <-d.quitChan:
			return
		default:
		}
		var res *response
		select {
		case <-d.quitChan:
			return
		case pktResponse := <-d.responseInput:
			if pktResponse == nil {
				return
			}
			common.Log.Debugf("Handling packet on device %d", d.id)
			seq := pktResponse.Result.GetSequence()
			res, ok = d.getSeq(seq)
			if !ok {
				common.Log.Warnf("Couldn't find requestor for seq %d on device %d", seq, d.id)
				continue
			}
			common.Log.Debugf("Returning seq %d to caller on device %d", seq, d.id)
			res.wg.Add(1)
			go func() {
				var timeout <-chan time.Time

				if d.timeout == nil || *d.timeout == 0 {
					timeout = make(<-chan time.Time)
				} else {
					timeout = time.After(*d.timeout)
				}
				select {
				case res.ch <- pktResponse:
					res.wg.Done()
				case <-timeout:
					res.wg.Done()
					common.Log.Warnf("Timeout returning seq %d to caller on device %d", seq, d.id)
				case <-res.done:
					res.wg.Done()
					select {
					case <-d.quitChan:
					default:
						d.delSeq(seq)
					}
				}
			}()
		}
	}
}

func (d *Device) addSeq() (seq uint8, res *response) {
	d.Lock()
	defer d.Unlock()
	d.sequence++
	if d.sequence == 0 {
		d.sequence++
	}
	seq = d.sequence
	res = &response{
		ch:   make(packet.Chan),
		done: make(doneChan),
	}
	d.responseMap[seq] = res

	return seq, res
}

func (d *Device) getSeq(seq uint8) (res *response, ok bool) {
	d.RLock()
	defer d.RUnlock()
	res, ok = d.responseMap[seq]

	return res, ok
}

func (d *Device) delSeq(seq uint8) {
	res, ok := d.getSeq(seq)
	if !ok {
		return
	}
	d.Lock()
	defer d.Unlock()
	close(res.ch)
	delete(d.responseMap, seq)
}

func New(addr *net.UDPAddr, requestSocket *net.UDPConn, timeout *time.Duration, retryInterval *time.Duration, reliable bool, pkt *packet.Packet) (*Device, error) {
	d := &Device{}
	d.init(addr, requestSocket, timeout, retryInterval, reliable)

	if pkt != nil {
		d.id = pkt.Target
		service := &stateService{}

		if err := pkt.DecodePayload(service); err != nil {
			return nil, err
		}

		d.address.Port = int(service.Port)
	}

	go d.handler()

	return d, nil
}
