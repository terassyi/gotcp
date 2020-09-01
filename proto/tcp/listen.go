package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/util"
)

type Listener struct {
	inner    *Tcp // TODO どうにかする
	synQueue chan AddressedPacket
	tcb      *controlBlock
}

func (t *Tcp) Listen(addr string, port int) (*Listener, error) {
	a, err := ipv4.StringToIPAddress(addr)
	if err != nil {
		return nil, err
	}
	return t.listen(a, port)
}

func (t *Tcp) listen(addr *ipv4.IPAddress, port int) (*Listener, error) {
	tcb, err := t.bind(port)
	if err != nil {
		return nil, err
	}
	listener := &Listener{
		inner: t,
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
	syn, ok := <-l.inner.SynQueue
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

	synAck, err := tcp.Build(uint16(l.tcb.peer.Port), uint16(l.tcb.peer.PeerPort),
		util.GetRandomUint32(), syn.Packet.Header.Sequence+1,
		tcp.SYN|tcp.ACK,
		syn.Packet.Header.WindowSize, 0, nil)
	if err != nil {
		return err
	}
	l.inner.enqueue(l.tcb.peer.PeerAddr, synAck)
	l.tcb.state = SYN_RECVD

	return nil
}
