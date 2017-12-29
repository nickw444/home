package packet

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"net"

	"github.com/lunixbochs/struc"
)

const checksumLengthBytes = 16

// See https://github.com/OpenMiHome/mihome-binary-protocol/blob/master/doc/PROTOCOL.md for
// documentation
type Header struct {
	Magic    uint16
	Length   uint16
	F1       uint32 // Unknown field.
	DeviceID uint32
	Stamp    uint32
	Checksum []byte `struc:"[16]byte"`
}

// Meta provides (optional) additional context about incoming packets.
type Meta struct {
	DecodeTime time.Time
	Addr       *net.UDPAddr
}

type Packet struct {
	Meta   Meta
	Header Header
	Data   []byte
}

func (p *Packet) Serialize() []byte {
	var buf bytes.Buffer
	err := struc.Pack(&buf, &p.Header)
	if err != nil {
		panic(err)
	}

	buf.Write(p.Data)
	return buf.Bytes()
}

func (p *Packet) CalcChecksum() ([]byte, error) {
	h := md5.New()
	_, err := h.Write(p.Serialize())
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (p *Packet) WriteChecksum() error {
	checksum, err := p.CalcChecksum()
	if err != nil {
		return err
	}

	p.Header.Checksum = checksum
	return nil
}

func (p *Packet) DataLength() int {
	return len(p.Data)
}

func (p *Packet) Verify(deviceToken []byte) error {
	var tmpPacket Packet
	tmpPacket = *p
	tmpPacket.Header.Checksum = deviceToken

	calculated, err := tmpPacket.CalcChecksum()
	if err != nil {
		return err
	}

	if !bytes.Equal(calculated, p.Header.Checksum) {
		return fmt.Errorf("Checksum could not be verified. Expected %s, got %s.",
			hex.EncodeToString(calculated), hex.EncodeToString(p.Header.Checksum))
	}
	return nil
}

func Decode(data []byte, addr *net.UDPAddr) (*Packet, error) {
	meta := Meta{DecodeTime: time.Now(), Addr: addr}
	header := Header{}
	struc.Unpack(bytes.NewBuffer(data[:32]), &header)

	p := &Packet{
		Meta:   meta,
		Header: header,
		Data:   data[32:],
	}
	return p, nil
}

// New creates a new packet
func New(deviceId uint32, deviceToken []byte, stamp uint32, data []byte) *Packet {
	header := Header{
		Magic:    0x2131,
		Length:   uint16(32 + len(data)),
		F1:       0x0,
		DeviceID: deviceId,
		Stamp:    stamp,
		Checksum: deviceToken,
	}

	p := &Packet{
		Header: header,
		Data:   data,
	}
	return p
}

func NewHello() *Packet {
	checksum := bytes.Repeat([]byte{0xff}, 16)
	return &Packet{
		Header: Header{
			Magic:    0x2131,
			Length:   0x0020,
			F1:       0xffffffff,
			DeviceID: 0xffffffff,
			Stamp:    0xffffffff,
			Checksum: checksum,
		},
	}
}
