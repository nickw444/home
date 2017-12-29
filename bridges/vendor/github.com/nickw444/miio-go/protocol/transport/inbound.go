package transport

import (
	"net"

	"github.com/nickw444/miio-go/protocol/packet"
)

// An inbound transport is a channel'ed abstraction around a net.UDPConn.
// Provides an abstraction around inbound packets on the network to allow
// using the existing protocol implementation without a miIO network handy.
// Consumers of this interface should never close the underlying UDP
// connection without first calling Stop().
type Inbound interface {
	Packets() <-chan *packet.Packet
	Stop() error
}

type inbound struct {
	socket   InboundConn
	packets  chan *packet.Packet
	quitChan chan struct{}
	stopped  bool
}

// InboundConn is an abstraction around net.UDPConn to allow
// mocking during tests.
type InboundConn interface {
	ReadFromUDP(b []byte) (int, *net.UDPAddr, error)
}

func NewInbound(socket InboundConn) Inbound {
	return newInbound(socket)
}

func newInbound(socket InboundConn) *inbound {
	i := &inbound{
		socket:   socket,
		packets:  make(chan *packet.Packet),
		quitChan: make(chan struct{}),
		stopped:  false,
	}
	go i.reader()
	return i
}

// A goroutine that continuously pulls data from the given UDP socket
// and decodes inbound packets.
func (i *inbound) reader() {
	for {
		select {
		case <-i.quitChan:
			return
		default:
			buf := make([]byte, 1024)
			n, addr, err := i.socket.ReadFromUDP(buf)

			if i.stopped {
				// No need to process this packet as we have been stopped.
				return
			}

			if err != nil {
				// TODO NW remove panic
				panic(err)
				continue
			}

			pkt, err := packet.Decode(buf[:n], addr)
			if err != nil {
				// TODO NW remove panic
				panic(err)
				continue
			}

			i.packets <- pkt
		}
	}
}

func (i *inbound) Packets() <-chan *packet.Packet {
	return i.packets
}

func (i *inbound) Stop() error {
	close(i.quitChan)
	i.stopped = true
	return nil
}
