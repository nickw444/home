package device

import "github.com/pdf/golifx/protocol/v2/packet"

type Location struct {
	// Location is a Group
	Group
}

func NewLocation(pkt *packet.Packet) (*Location, error) {
	l := new(Location)
	l.init()
	err := l.Parse(pkt)
	return l, err
}
