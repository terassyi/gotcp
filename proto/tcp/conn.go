package tcp

import (
	"github.com/terassyi/gotcp/packet/ipv4"
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
	peerAddr, err := ipv4.StringToIPAddress(addr)
	if err != nil {
		return nil, err
	}
	peer, err := t.Table.Add(peerAddr, peerport, 0)
	if err != nil {
		return nil, err
	}
	conn, err := newConn(peer)
	if err != nil {
		return nil, err
	}
	conn.establish()
	return nil, nil
}

func (conn *Conn) establish() error {
	// tcp active open
	p, err := conn.activeOpen()
	if err != nil {
		return err
	}
	// send syn
	conn.inner.enqueue(conn.peer.PeerAddr, p)

	// wait to receive syn|ack packet

}
