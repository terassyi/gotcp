package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/packet/tcp"
)

type Listener struct {
	inner *Tcp // TODO どうにかする
	queue chan AddressedPacket
	tcb   *controlBlock
}

func (t *Tcp) Listen(addr string, port int) (*Listener, error) {
	a, err := ipv4.StringToIPAddress(addr)
	if err != nil {
		return nil, err
	}
	fmt.Println("[info] start to listen")
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	l, err := t.listen(a, port)
	if err != nil {
		return nil, err
	}
	t.listeners[port] = l
	return l, nil
}

func (t *Tcp) listen(addr *ipv4.IPAddress, port int) (*Listener, error) {
	tcb, err := t.bind(port)
	if err != nil {
		return nil, err
	}
	listener := &Listener{
		inner: t,
		queue: make(chan AddressedPacket, 100),
		tcb:   tcb,
	}

	return listener, nil
}

func (t *Tcp) bind(port int) (*controlBlock, error) {
	// TODO もし並列にコネクション貼るならPeerのポインタ渡すのは良くないかも
	peer, err := t.Table.Bind(port)
	if err != nil {
		return nil, err
	}
	cb := NewControlBlock(peer)
	fmt.Println("[info] bind port=", port)
	if err := cb.passiveOpen(); err != nil {
		return nil, err
	}
	return cb, nil
}

func (l *Listener) Accept() (*Conn, error) {
	// wait syn packet recv from client
	if err := l.establish(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (l *Listener) establish() error {
	syn, ok := <-l.queue
	if !ok {
		return fmt.Errorf("failed to recv syn from syn queue")
	}
	if syn.Packet.Header.DestinationPort != uint16(l.tcb.peer.Port) {
		// send reset
		rep, err := tcp.Build(syn.Packet.Header.DestinationPort, syn.Packet.Header.SourcePort,
			0, 0, tcp.RST, 0, 0, nil)
		if err != nil {
			return err
		}
		l.inner.enqueue(syn.Address, rep)
		return fmt.Errorf("an packet to unbinded port is recieved")
	}

	l.tcb.peer.PeerAddr = syn.Address
	l.tcb.peer.PeerPort = int(syn.Packet.Header.SourcePort)

	// update recv sequence
	l.tcb.Rcv.NXT = syn.Packet.Header.Sequence + 1
	l.tcb.Rcv.IRS = syn.Packet.Header.Sequence
	l.tcb.Snd.ISS = Random()
	l.tcb.Snd.NXT = l.tcb.Snd.ISS + 1
	l.tcb.Snd.UNA = l.tcb.Snd.ISS

	synAck, err := tcp.Build(uint16(l.tcb.peer.Port), uint16(l.tcb.peer.PeerPort),
		l.tcb.Snd.ISS, l.tcb.Rcv.NXT,
		tcp.SYN|tcp.ACK,
		syn.Packet.Header.WindowSize, 0, nil)
	if err != nil {
		return err
	}
	opTimeStamp := syn.Packet.Option.TimeStamp()
	synAck.AddOption(tcp.Options{tcp.MaxSegmentSize(1460), tcp.SACKPermitted{}, tcp.WindowScale(7), opTimeStamp.Exchange()})
	l.inner.enqueue(l.tcb.peer.PeerAddr, synAck)
	l.tcb.state = SYN_RECVD
	fmt.Println("[info] transmission control block state is SYN_RECVD")
	// wait ack
	ack, ok := <-l.queue
	if !ok {
		return fmt.Errorf("failed to recv syn from syn queue")
	}
	// if not ack
	if !ack.Packet.Header.OffsetControlFlag.ControlFlag().Ack() {
		rep, err := tcp.Build(syn.Packet.Header.DestinationPort, syn.Packet.Header.SourcePort,
			0, 0, tcp.RST, 0, 0, nil)
		if err != nil {
			return err
		}
		l.inner.enqueue(syn.Address, rep)
	}
	if l.tcb.Snd.UNA <= ack.Packet.Header.Ack && ack.Packet.Header.Ack <= l.tcb.Snd.NXT {
		fmt.Println("[info] status move to ESTABLISHED")
		l.tcb.state = ESTABLISHED
	} else {
		rep, err := tcp.Build(syn.Packet.Header.DestinationPort, syn.Packet.Header.SourcePort,
			0, 0, tcp.RST, 0, 0, nil)
		if err != nil {
			return err
		}
		l.inner.enqueue(syn.Address, rep)
	}
	return nil
}
