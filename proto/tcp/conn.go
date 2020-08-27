package tcp

import (
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/proto/port"
)

type Conn struct {
	*controlBlock
	Peer *port.Table
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
	cb := NewControlBlock(peer)

	p, err := cb.activeOpen()
	if err != nil {
		return nil, err
	}
	t.enqueue(peerAddr, p)
	return nil, nil
}
