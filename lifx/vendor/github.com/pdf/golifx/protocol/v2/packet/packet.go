// Package packet implements a LIFX LAN protocol version 2 packet.
//
// This package is not designed to be accessed by end users, all interaction
// should occur via the Client in the golifx package.
package packet

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/lunixbochs/struc"

	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol/v2/shared"
)

const (
	headerBytes = 36
)

var (
	ClientID uint32
)

type Response struct {
	Result Packet
	Error  error
}

type Packet struct {
	Frame
	FrameAddress
	ProtocolHeader

	payload     []byte
	destination *net.UDPAddr
	socket      *net.UDPConn
	mutex       sync.RWMutex
}

type Chan chan Response

type Frame struct {
	Size uint16 `struc:"little"` // Size of entire message in bytes including this field
	// Origin (2-bit): Message origin indicator: must be zero (0)
	// Tagged (1-bit): Determines usage of Frame Address *target* field
	// Addressable (1-bit): Message includes a target address: must be one (1)
	// Protocol (12-bit): Protocol number: must be 1024 (decimal)
	OriginTaggedAddressableProtocol uint16 `struct:"little"`
	Source                          uint32 `struc:"little"` // Source identifier: unique value set by the client, used by responses
}

type FrameAddress struct {
	Target   uint64   `struc:"little"` // Device address (MAC address) or zero (0) means all devices
	Reserved [6]uint8 `struc:"little"`
	// Reserved(6-bit)
	// AckRequired(1-bit): Acknowledgement message required
	// ResRequired(1-bit): Response message required
	ReservedAckRequiredResRequired uint8 `struc:"little"`
	Sequence                       uint8 `struc:"little"`
}

type ProtocolHeader struct {
	Reserved0 uint64 `struc:"little"` // Reserved
	Type      uint16 `struc:"little"` // Message type determines the payload being used
	Reserved1 uint16 `struc:"little"` // Reserved
}

func (f *Frame) SetSize(size uint16) {
	f.Size = size
}

func (f *Frame) SetTagged(tagged bool) {
	if tagged {
		f.OriginTaggedAddressableProtocol = uint16(setBits(uint64(f.OriginTaggedAddressableProtocol), 16, 2, 1, 1))
	} else {
		f.OriginTaggedAddressableProtocol = uint16(setBits(uint64(f.OriginTaggedAddressableProtocol), 16, 2, 1, 0))
	}
}

func (f *Frame) GetTagged() bool {
	return getBits(uint64(f.OriginTaggedAddressableProtocol), 16, 2, 1) != 0
}

func (f *Frame) SetAddressable(addressable bool) {
	if addressable {
		f.OriginTaggedAddressableProtocol = uint16(setBits(uint64(f.OriginTaggedAddressableProtocol), 16, 3, 1, 1))
	} else {
		f.OriginTaggedAddressableProtocol = uint16(setBits(uint64(f.OriginTaggedAddressableProtocol), 16, 3, 1, 0))
	}
}

func (f *Frame) GetAddressable() bool {
	return getBits(uint64(f.OriginTaggedAddressableProtocol), 16, 3, 1) != 0
}

func (f *Frame) SetProtocol(protocol uint16) {
	f.OriginTaggedAddressableProtocol = uint16(setBits(uint64(f.OriginTaggedAddressableProtocol), 16, 4, 12, uint64(protocol)))
}

func (f *Frame) GetProtocol() uint16 {
	return uint16(getBits(uint64(f.OriginTaggedAddressableProtocol), 16, 4, 12))
}

func (f *Frame) SetSource(source uint32) {
	f.Source = source
}

func (f *Frame) GetSource() uint32 {
	return f.Source
}

func (f *FrameAddress) SetTarget(target uint64) {
	f.Target = target
}

func (f *FrameAddress) GetTarget() uint64 {
	return f.Target
}

func (f *FrameAddress) SetAckRequired(ackRequired bool) {
	if ackRequired {
		f.ReservedAckRequiredResRequired = uint8(setBits(uint64(f.ReservedAckRequiredResRequired), 8, 6, 1, 1))
	} else {
		f.ReservedAckRequiredResRequired = uint8(setBits(uint64(f.ReservedAckRequiredResRequired), 8, 6, 1, 0))
	}
}

func (f *FrameAddress) GetAckRequired() bool {
	return getBits(uint64(f.ReservedAckRequiredResRequired), 8, 6, 1) != 0
}

func (f *FrameAddress) SetResRequired(resRequired bool) {
	if resRequired {
		f.ReservedAckRequiredResRequired = uint8(setBits(uint64(f.ReservedAckRequiredResRequired), 8, 7, 1, 1))
	} else {
		f.ReservedAckRequiredResRequired = uint8(setBits(uint64(f.ReservedAckRequiredResRequired), 8, 7, 1, 1))
	}
}

func (f *FrameAddress) GetResRequired() bool {
	return getBits(uint64(f.ReservedAckRequiredResRequired), 8, 7, 1) != 0
}

func (f *FrameAddress) SetSequence(sequence uint8) {
	f.Sequence = sequence
}

func (f *FrameAddress) GetSequence() uint8 {
	//sequence := uint64(math.MaxUint8)
	//sequence &= f.ReservedAckRequiredResRequiredSequence
	return f.Sequence
}

func (h *ProtocolHeader) GetType() shared.Message {
	return shared.Message(h.Type)
}

func (h *ProtocolHeader) SetType(msgType shared.Message) {
	h.Type = uint16(msgType)
}

func (p *Packet) Write() error {
	var (
		err     error
		byteArr []byte
		size    []byte
		sizeBuf *bytes.Buffer
	)

	// Encode
	p.Frame.Size, byteArr, err = p.encode()
	if err != nil {
		return err
	}
	sizeBuf = new(bytes.Buffer)
	if err := binary.Write(sizeBuf, binary.LittleEndian, p.Frame.Size); err != nil {
		return err
	}
	size = sizeBuf.Bytes()
	// Set size
	for i, b := range size {
		byteArr[i] = b
	}

	common.Log.Debugf("Writing packet data: source %v, type %v, sequence %v, target %v, protocol %v, tagged %v, resRequired %v, ackRequired %v, frameOrigin %b: %+v", p.GetSource(), p.GetType(), p.GetSequence(), p.GetTarget(), p.GetProtocol(), p.GetTagged(), p.GetResRequired(), p.GetAckRequired(), p.Frame.OriginTaggedAddressableProtocol, *p)

	// Write the packet
	_, err = p.socket.WriteToUDP(byteArr, p.destination)
	return err
}

func (p *Packet) SetPayload(payload interface{}) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, payload); err != nil {
		return err
	}
	p.mutex.Lock()
	p.payload = buf.Bytes()
	p.mutex.Unlock()
	return nil
}

func (p *Packet) GetPayload() []byte {
	return p.payload
}

func (p *Packet) DecodePayload(dest interface{}) error {
	p.mutex.RLock()
	if len(p.payload) == 0 {
		p.mutex.RUnlock()
		return common.ErrProtocol
	}
	p.mutex.RUnlock()
	p.mutex.Lock()
	err := binary.Read(bytes.NewBuffer(p.payload), binary.LittleEndian, dest)
	p.mutex.Unlock()
	return err
}

func (p *Packet) encode() (length uint16, byteArr []byte, err error) {
	buf := new(bytes.Buffer)

	// Pack the packet
	p.mutex.RLock()
	err = binary.Write(buf, binary.LittleEndian, p.Frame)
	if err != nil {
		p.mutex.RUnlock()
		return
	}
	err = binary.Write(buf, binary.LittleEndian, p.FrameAddress)
	if err != nil {
		p.mutex.RUnlock()
		return
	}
	err = binary.Write(buf, binary.LittleEndian, p.ProtocolHeader)
	if err != nil {
		p.mutex.RUnlock()
		return
	}
	//err = struc.Pack(buf, p)
	p.mutex.RUnlock()
	if err != nil {
		return
	}

	// Append the packed payload
	if len(p.payload) > 0 {
		p.mutex.RLock()
		_, err = buf.Write(p.payload)
		p.mutex.RUnlock()
		if err != nil {
			return
		}
	}

	return uint16(buf.Len()), buf.Bytes(), nil
}

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
	ClientID = rand.Uint32()
}

func New(destination *net.UDPAddr, socket *net.UDPConn) *Packet {
	p := &Packet{
		destination: destination,
		socket:      socket,
	}
	p.Frame.SetSource(ClientID)
	p.Frame.SetAddressable(true) // Must be true
	p.Frame.SetProtocol(1024)    // Must be 1024
	return p
}

func Decode(buf []byte) (*Packet, error) {
	pkt := new(Packet)
	err := struc.Unpack(bytes.NewBuffer(buf), pkt)
	if err != nil {
		return pkt, err
	}
	pkt.payload = buf[headerBytes:]
	return pkt, nil
}

func bits(n uint8) uint64 {
	return (1 << n) - 1
}

func bitMask(endLSB, startLSB uint8) uint64 {
	return bits(endLSB) &^ bits(startLSB)
}

func rangeToLSB(size, index, length uint8) (endLSB, startLSB uint8) {
	startLSB = size - index - length
	endLSB = startLSB + length
	return endLSB, startLSB
}

func setBits(input uint64, size, index, length uint8, value uint64) uint64 {
	endLSB, startLSB := rangeToLSB(size, index, length)
	return input&^bitMask(endLSB, startLSB) | (value << startLSB)
}

func getBits(input uint64, size, index, length uint8) uint64 {
	endLSB, startLSB := rangeToLSB(size, index, length)
	return (input & bitMask(endLSB, startLSB)) >> startLSB
}
