package tcp

import (
	"github.com/terassyi/gotcp/proto/port"
)

type Conn struct {
	*controlBlock
	Peer  *port.Peer
	queue chan AddressedPacket
	inner *Tcp
}

func newConn(peer *port.Peer) (*Conn, error) {
	return &Conn{
		controlBlock: NewControlBlock(peer),
		Peer:         peer,
		queue:        make(chan AddressedPacket, 100),
	}, nil
}

func (t *Tcp) Dial(addr string, peerport int) (*Conn, error) {
	return t.doDial(addr, peerport)
}

func (t *Tcp) doDial(addr string, peerport int) (*Conn, error) {
	dialer, err := t.dial(addr, peerport)
	if err != nil {
		return nil, err
	}
	return dialer.getConnection()
}
