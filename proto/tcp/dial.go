package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/proto/port"
)

type dialer struct {
	tcb   *controlBlock
	peer  *port.Peer
	queue chan AddressedPacket
	inner *Tcp
}

func newDialer(inner *Tcp, peer *port.Peer) (*dialer, error) {
	return &dialer{
		tcb:   NewControlBlock(peer),
		queue: make(chan AddressedPacket, 100),
		inner: inner,
	}, nil
}

func (t *Tcp) dial(addr string, peerport int) (*dialer, error) {
	peerAddr, err := ipv4.StringToIPAddress(addr)
	if err != nil {
		return nil, err
	}
	peer, err := t.Table.Add(peerAddr, peerport, 0)
	if err != nil {
		return nil, err
	}
	d, err := newDialer(t, peer)
	if err != nil {
		return nil, err
	}
	if err := d.establish(); err != nil {
		return nil, err
	}
	t.dialers[peerport] = d
	return d, nil
}

func (d *dialer) establish() error {
	// tcp active open
	p, err := d.tcb.activeOpen()
	if err != nil {
		return err
	}
	d.inner.enqueue(d.peer.PeerAddr, p)

	// wait to receive syn|ack packet
	synAck, ok := <-d.queue
	if !ok {
		fmt.Println("failed to recv syn from syn queue")
	}
	if !synAck.Packet.Header.OffsetControlFlag.ControlFlag().Syn() || !synAck.Packet.Header.OffsetControlFlag.ControlFlag().Ack() {
		rep, err := tcp.Build(synAck.Packet.Header.DestinationPort, synAck.Packet.Header.SourcePort,
			0, 0, tcp.RST, 0, 0, nil)
		if err != nil {
			return err
		}
		d.inner.enqueue(synAck.Address, rep)
		return fmt.Errorf("received packet is not set syn|ack.")
	}
	// handle syn|ack
	// This step should be reached only if the ACK is ok, or there is no ACK, and it the segment did not contain a RST.
	d.tcb.Rcv.NXT = synAck.Packet.Header.Sequence + 1
	d.tcb.Rcv.IRS = synAck.Packet.Header.Sequence
	d.tcb.Snd.UNA = synAck.Packet.Header.Ack
	if d.tcb.Snd.ISS < d.tcb.Snd.UNA {
		d.tcb.ESTABLISHED()
		fmt.Println("[info] transmission control block state is ESTABLISHED")
		ack, err := tcp.Build(
			uint16(d.tcb.peer.Port), uint16(d.peer.PeerPort),
			d.tcb.Snd.NXT, d.tcb.Rcv.NXT,
			tcp.ACK,
			synAck.Packet.Header.WindowSize, 0, nil)
		if err != nil {
			return err
		}
		// send ack packet
		d.inner.enqueue(d.tcb.peer.PeerAddr, ack)
	}
	return nil
}
