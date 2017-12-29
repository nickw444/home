package transport

import (
	"net"

	"github.com/nickw444/miio-go/protocol/packet"
)

type Conn interface {
	InboundConn
	OutboundConn
	Close() error
}

type Transport interface {
	Inbound() Inbound
	NewOutbound(crypto packet.Crypto, dest net.Addr) Outbound
	Close() error
}

type transport struct {
	inbound   Inbound
	outbounds []Outbound
	socket    Conn
}

func NewTransport(socket Conn) Transport {
	return &transport{
		socket: socket,
	}
}

func (t *transport) Inbound() Inbound {
	if t.inbound == nil {
		t.inbound = NewInbound(t.socket)
	}
	return t.inbound
}

func (t *transport) NewOutbound(crypto packet.Crypto, dest net.Addr) Outbound {
	o := NewOutbound(crypto, dest, t.socket)
	t.outbounds = append(t.outbounds, o)
	return o
}

func (t *transport) Close() error {
	err := t.inbound.Stop()
	if err != nil {
		return err
	}
	err = t.socket.Close()
	return err
}
