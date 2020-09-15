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
	d := &dialer{
		tcb:   NewControlBlock(peer),
		peer:  peer,
		queue: make(chan AddressedPacket, 100),
		inner: t,
	}
	d.tcb.rcv.WND = 1024
	t.dialers[peer.Port] = d
	if err := d.establish(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *dialer) establish() error {
	// tcp active open
	d.tcb.mutex.RLock()
	defer d.tcb.mutex.RUnlock()
	p, err := d.tcb.activeOpen()
	if err != nil {
		return err
	}
	d.inner.enqueue(d.peer.PeerAddr, p)
	//d.tcb.snd.NXT += 1
	d.tcb.showSeq()
	fmt.Println("[info] waiting for syn ack packet")
	// wait to receive syn|ack packet
	synAck, ok := <-d.queue
	if !ok {
		fmt.Println("failed to recv syn from syn queue")
	}
	//d.tcb.rcv.NXT += 1
	d.tcb.showSeq()
	fmt.Println("[debug] received syn ack packet")
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
	fmt.Println("[debug] handling syn ack packet")
	// This step should be reached only if the ACK is ok, or there is no ACK, and it the segment did not contain a RST.
	d.tcb.rcv.NXT = synAck.Packet.Header.Sequence + 1
	d.tcb.rcv.IRS = synAck.Packet.Header.Sequence
	d.tcb.snd.UNA = synAck.Packet.Header.Ack
	if d.tcb.snd.ISS < d.tcb.snd.UNA {
		d.tcb.ESTABLISHED()
		fmt.Println("[info] transmission control block state is ESTABLISHED")
		ack, err := tcp.Build(
			uint16(d.tcb.peer.Port), uint16(d.peer.PeerPort),
			d.tcb.snd.NXT, d.tcb.rcv.NXT,
			tcp.ACK,
			synAck.Packet.Header.WindowSize, 0, nil)
		if err != nil {
			return err
		}
		//d.tcb.snd.NXT += 1
		// send ack packet
		d.inner.enqueue(d.tcb.peer.PeerAddr, ack)
		d.tcb.showSeq()
		fmt.Println("[info] finished 3 way handshake")
		return nil
	}
	d.tcb.snd.Show()
	return fmt.Errorf("invalid tcb")
}

func (d *dialer) getConnection() (*Conn, error) {
	conn := &Conn{
		tcb:        d.tcb,
		Peer:       d.peer,
		queue:      make(chan AddressedPacket, 100),
		closeQueue: make(chan AddressedPacket, 1),
		rcvBuffer:  make([]byte, window),
		readyQueue: make(chan []byte, 10),
		inner:      d.inner,
	}
	conn.pushFlag = true
	// entry connection list
	d.inner.connections[conn.Peer.Port] = conn
	// delete dialer from dialer list
	delete(d.inner.dialers, conn.Peer.Port)
	return conn, nil
}
