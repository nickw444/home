package packet

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

var payload []byte
var deviceToken []byte
var packet *Packet

func setUp(t *testing.T) func(t *testing.T) {
	payload = []byte("Hello World")
	deviceToken = bytes.Repeat([]byte{0xFF, 0x00}, checksumLengthBytes/2)
	packet = New(0xAAAABBBB, deviceToken, 0xCCCCDDDD, payload)
	return func(t *testing.T) {
		payload = nil
		deviceToken = nil
		packet = nil
	}
}

// Ensure a known packet decodes and re-serializes to the same value
func TestDecode(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	data := packet.Serialize()
	newPkt, err := Decode(data, &net.UDPAddr{})
	assert.NoError(t, err)

	// Ensure the new packet serializes the same
	newData := newPkt.Serialize()
	assert.Equal(t, data, newData)
}

// Ensure that serialize orders fields correctly
func TestPacket_Serialize(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	data := packet.Serialize()
	assert.Equal(t, []byte{0x21, 0x31}, data[0:2])
	assert.Equal(t, uint16(len(payload)+32), binary.BigEndian.Uint16(data[2:4]))
	assert.Equal(t, uint32(0), binary.BigEndian.Uint32(data[4:8]))
	assert.Equal(t, uint32(0xAAAABBBB), binary.BigEndian.Uint32(data[8:12]))
	assert.Equal(t, uint32(0xCCCCDDDD), binary.BigEndian.Uint32(data[12:16]))
	assert.Equal(t, deviceToken, data[16:32])
}

// Ensure that CalcChecksum outputs a known good value.
func TestPacket_CalcChecksum(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	chk, err := packet.CalcChecksum()
	assert.NoError(t, err)

	assert.Equal(t, []byte{0x41, 0x7a, 0x35, 0xb4, 0x21, 0x5c, 0x64, 0xad, 0xd5, 0xe0, 0xcd, 0x3f, 0x51, 0x47, 0xf5, 0xc2}, chk)
}

// Ensure that the written checksum is the same value that is calculated.
func TestPacket_WriteChecksum(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	packet.Header.Checksum = bytes.Repeat([]byte{0xFF}, checksumLengthBytes)

	chk, err := packet.CalcChecksum()
	assert.NoError(t, err)

	err = packet.WriteChecksum()
	assert.NoError(t, err)
	assert.Equal(t, chk, packet.Header.Checksum)

}

// Ensure that the packet data length returned is expected.
func TestPacket_DataLength(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	assert.Equal(t, len(payload), packet.DataLength())
}

// Test verification with a malformed packet checksum.
func TestPacket_VerifyFail1(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	err := packet.WriteChecksum()
	assert.NoError(t, err)

	// Mutate the checksum
	packet.Header.Checksum[0]++

	err = packet.Verify(deviceToken)
	assert.NotNil(t, err)
}

// Test verification with a malformed packet checksum.
func TestPacket_VerifyFail2(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	err := packet.WriteChecksum()
	assert.NoError(t, err)

	// Mutate the data
	packet.Data[0]++

	err = packet.Verify(deviceToken)
	assert.NotNil(t, err)
}

// Test verification with a known good packet.
func TestPacket_VerifySuccess(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	err := packet.WriteChecksum()
	assert.NoError(t, err)

	err = packet.Verify(deviceToken)
	assert.NoError(t, err)
}

// Ensure that calls to Verify does not mutate the packet header.
func TestPacket_VerifyNoMutation(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	packet.WriteChecksum()

	before := packet.Serialize()
	packet.Verify(deviceToken)

	after := packet.Serialize()
	assert.Equal(t, before, after)
}
