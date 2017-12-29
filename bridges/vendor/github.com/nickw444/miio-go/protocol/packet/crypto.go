package packet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
)

type Crypto interface {
	VerifyPacket(pkt *Packet) error
	Decrypt(data []byte) ([]byte, error)
	Encrypt(data []byte) ([]byte, error)
	NewPacket(data []byte) (*Packet, error)
}

type crypto struct {
	iv           []byte
	key          []byte
	deviceId     uint32
	deviceToken  []byte
	initialStamp uint32
	stampTime    time.Time
	clock        clock.Clock
}

func NewCrypto(deviceID uint32, deviceToken []byte, initialStamp uint32, stampTime time.Time, clock clock.Clock) (
	Crypto, error) {

	hash := md5.New()
	_, err := hash.Write(deviceToken)
	if err != nil {
		return nil, err
	}
	key := hash.Sum(nil)

	hash = md5.New()
	_, err = hash.Write(key)
	if err != nil {
		return nil, err
	}
	_, err = hash.Write(deviceToken)
	if err != nil {
		return nil, err
	}
	iv := hash.Sum(nil)

	return &crypto{
		deviceId:     deviceID,
		deviceToken:  deviceToken,
		initialStamp: initialStamp,
		stampTime:    stampTime,
		clock:        clock,

		iv:  iv,
		key: key,
	}, nil
}

func (c *crypto) VerifyPacket(pkt *Packet) error {
	// Verify the checksum.
	return pkt.Verify(c.deviceToken)
}

func (c *crypto) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCBCDecrypter(block, c.iv)
	decrypted := make([]byte, len(data))
	stream.CryptBlocks(decrypted, data)

	return c.pkcs5Unpad(decrypted, block.BlockSize())
}

func (c *crypto) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	data = c.pkcs5Pad(data, block.BlockSize())
	stream := cipher.NewCBCEncrypter(block, c.iv)

	encrypted := make([]byte, len(data))
	stream.CryptBlocks(encrypted, []byte(data))
	return encrypted, nil
}

// Pad using PKCS5 padding scheme.
func (c *crypto) pkcs5Pad(data []byte, blockSize int) []byte {
	length := len(data)
	padLength := (blockSize - (length % blockSize))
	pad := bytes.Repeat([]byte{byte(padLength)}, padLength)
	return append(data, pad...)
}

// Unpad using PKCS5 padding scheme.
func (c *crypto) pkcs5Unpad(data []byte, blockSize int) ([]byte, error) {
	srcLen := len(data)
	paddingLen := int(data[srcLen-1])
	if paddingLen >= srcLen || paddingLen > blockSize {
		return nil, fmt.Errorf("Padding size error whilst decrypting payload.")
	}
	return data[:srcLen-paddingLen], nil
}

func (c *crypto) getStamp() uint32 {
	return uint32(c.clock.Now().Sub(c.stampTime).Seconds()) + c.initialStamp
}

func (c *crypto) NewPacket(data []byte) (*Packet, error) {
	encrypted, err := c.Encrypt(data)
	if err != nil {
		return nil, err
	}

	stamp := c.getStamp()

	p := New(c.deviceId, c.deviceToken, stamp, encrypted)
	err = p.WriteChecksum()

	if err != nil {
		return nil, err
	}

	return p, nil
}
