package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/proto/port"
)

type Conn struct {
	tcb        *controlBlock
	Peer       *port.Peer
	queue      chan AddressedPacket
	closeQueue chan AddressedPacket
	rcvBuffer  []byte
	inner      *Tcp
}

func newConn(peer *port.Peer) (*Conn, error) {
	conn := &Conn{
		tcb:       NewControlBlock(peer),
		Peer:      peer,
		queue:     make(chan AddressedPacket, 100),
		rcvBuffer: make([]byte, 0, 1024),
	}
	conn.tcb.rcv.WND = 1024
	return conn, nil
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

func (c *Conn) Close() error {
	return c.activeClose()
}

func (c *Conn) activeClose() error {
	// close tcb
	if c.tcb.state != ESTABLISHED && c.tcb.state != CLOSE_WAIT && c.tcb.state != SYN_RECVD {
		return fmt.Errorf("invalid state")
	}
	if err := c.send(tcp.ACK|tcp.FIN, nil); err != nil {
		return err
	}
	c.tcb.finSend = true
	if c.tcb.state == SYN_RECVD || c.tcb.state == ESTABLISHED {
		fmt.Println("[info] transmission control block state is FIN-WAIT1")
		c.tcb.FIN_WAIT1()
		// wait ack of fin
		p, ok := <-c.closeQueue
		if !ok {
			return fmt.Errorf("failed to recieve ack of fin.")
		}
		if p.Packet.Header.OffsetControlFlag.ControlFlag().Fin() {
			// simultaneous close
			if err := c.send(tcp.ACK, nil); err != nil {
				return err
			}
			c.tcb.CLOSING()
			_, ok := <-c.closeQueue
			if !ok {
				return fmt.Errorf("failed to recieve ack of fin")
			}
			c.tcb.TIME_WAIT()
			c.tcb.startMSL()
			c.tcb.CLOSED()
			return nil
		}
		fmt.Println("[info] transmission control block state is FIN-WAIT2")
		c.tcb.FIN_WAIT2()
		// wait fin
		_, ok = <-c.closeQueue
		if !ok {
			return fmt.Errorf("failed to recieve ack of fin.")
		}
		if err := c.send(tcp.ACK, nil); err != nil {
			return err
		}
		c.tcb.TIME_WAIT()
		c.tcb.startMSL()
		c.tcb.CLOSED()

	} else {
		fmt.Println("[info] transmission control block state is LAST-ACK")
		c.tcb.CLOSE_WAIT()
	}
	return nil
}

func (c *Conn) passiveClose(fin AddressedPacket) error {
	// close tcb
	if c.tcb.state == CLOSED || c.tcb.state == LISTEN || c.tcb.state == SYN_SENT {
		// drop packet
		return nil
	}
	if c.tcb.state == SYN_RECVD || c.tcb.state == ESTABLISHED {
		fmt.Println("[info] transmission control block state is CLOSE-WAIT")
		c.tcb.CLOSE_WAIT()
	}
	if c.tcb.state == FIN_WAIT1 {
		if c.tcb.finSend {
			fmt.Println("[info] transmission control block state is TIME-WAIT")
			c.tcb.TIME_WAIT()
			// start 2MSL
			// delete TCB
		} else {
			fmt.Println("[info] transmission control block state is CLOSING")
			c.tcb.CLOSING()
		}
	}
	if c.tcb.state == FIN_WAIT2 {
		fmt.Println("[info] transmission control block state is TIME-WAIT")
		c.tcb.TIME_WAIT()
	}
	if c.tcb.state == CLOSING || c.tcb.state == LAST_ACK {
		// stay
	}
	if c.tcb.state == TIME_WAIT {
		// restart 2MSL
		c.tcb.startMSL()
		c.tcb.startMSL()
	}
	return nil
}

func (c *Conn) handle(packet AddressedPacket) error {
	// handle incoming segment
	fmt.Println("[info] handling connection, handling incoming segment.")
	// first check sequence number
	/*
	   Segment Receive  Test
	   Length  Window
	   ------- -------  -------------------------------------------

	      0       0     SEG.SEQ = RCV.NXT

	      0      >0     RCV.NXT =< SEG.SEQ < RCV.NXT+RCV.WND

	     >0       0     not acceptable

	     >0      >0     RCV.NXT =< SEG.SEQ < RCV.NXT+RCV.WND
	                 or RCV.NXT =< SEG.SEQ+SEG.LEN-1 < RCV.NXT+RCV.WND
	*/
	fmt.Println("[debug] (first) check sequence number")
	if c.tcb.rcv.WND == 0 || packet.Packet.Header.Sequence != c.tcb.rcv.NXT {
		//if err := c.send(tcp.ACK, nil); err != nil {
		//	return err
		//}
		fmt.Println("[debug] packet seq num = ", packet.Packet.Header.Sequence)
		fmt.Println("[debug] recv nxt = ", c.tcb.rcv.NXT)
		return fmt.Errorf("recieve window is zero")
	}

	// second check the RST bit,
	fmt.Println("[debug] (second) check RST")
	if packet.Packet.Header.OffsetControlFlag.ControlFlag().Rst() {
		c.tcb.CLOSED()
		return nil
	}
	// third check security and precedence
	fmt.Println("[debug] (third) check security and precedence nop")
	// TODO

	// fourth, check the SYN bit
	fmt.Println("[debug] (fourth) check SYN")
	if packet.Packet.Header.OffsetControlFlag.ControlFlag().Syn() {
		if err := c.send(tcp.RST, nil); err != nil {
			return err
		}
		c.tcb.CLOSED()
		return nil
	}
	// fifth check the ACK field
	fmt.Println("[debug] (fifth) check ACK")
	if packet.Packet.Header.OffsetControlFlag.ControlFlag().Ack() {
		switch c.tcb.state {
		case ESTABLISHED:
			fmt.Println("[debug] pass here")
			c.handleEstablished(packet)
		case FIN_WAIT1:
			c.handleEstablished(packet)
			if c.tcb.finSend {
				c.closeQueue <- packet
			}
			return nil
		case FIN_WAIT2:
			c.handleEstablished(packet)
			if len(c.tcb.retrans) == 0 {
				c.tcb.TIME_WAIT()
			}
		case CLOSE_WAIT:
			c.handleEstablished(packet)
		case CLOSING:
			c.handleEstablished(packet)
			if c.tcb.finSend {
				c.tcb.TIME_WAIT()
			}
		case LAST_ACK:
			// only reach here is when acknowledgement of my FIN
			if c.tcb.finSend {
				c.tcb.CLOSED()
			}
		case TIME_WAIT:
			// resend fin
			if err := c.send(tcp.FIN|tcp.ACK, nil); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("ack field is not set")
	}
	// sixth check the URG bit
	fmt.Println("[debug] (sixth) check URG nop")
	// TODO

	// seventh process the segment text
	fmt.Println("[debug] (seventh) process the segment text")
	switch c.tcb.state {
	case ESTABLISHED, FIN_WAIT1, FIN_WAIT2:
		if err := c.handleSegment(packet); err != nil {
			return err
		}
	default:
		// ignore the segment
	}
	// eighth check fin bit
	fmt.Println("[debug] (eight) check FIN")
	switch c.tcb.state {
	case CLOSED, LISTEN, SYN_SENT:
		fmt.Println("[debug] hoge~")
		// ignore
	case SYN_RECVD, ESTABLISHED:
		fmt.Println("[debug] reach handle fin phase")
		if err := c.handleFin(packet); err != nil {
			return err
		}
		c.tcb.CLOSE_WAIT()
	case FIN_WAIT1:
		fmt.Println("[debug] fuga")
		// ack
	case FIN_WAIT2:
		fmt.Println("[debug] reach passive close phase")
		c.closeQueue <- packet
		//if err := c.handleFin(packet); err != nil {
		if err := c.passiveClose(packet); err != nil {
			return err
		}
		c.tcb.TIME_WAIT()
	default:
		fmt.Println("[debug] stay ", c.tcb.state.String())
		// stay
	}
	return nil
}

func (c *Conn) handleEstablished(packet AddressedPacket) {
	if c.tcb.snd.UNA < packet.Packet.Header.Ack && packet.Packet.Header.Ack <= c.tcb.snd.NXT {
		c.tcb.snd.UNA = packet.Packet.Header.Ack

		// SND.WL1 < SEG.SEQ or (SND.WL1 = SEG.SEQ and SND.WL2 =< SEG.ACK)
		if c.tcb.snd.WL1 < packet.Packet.Header.Sequence || (c.tcb.snd.WL1 == packet.Packet.Header.Sequence && c.tcb.snd.WL2 <= packet.Packet.Header.Ack) {
			c.tcb.snd.WND = packet.Packet.Header.WindowSize
			c.tcb.snd.WL1 = packet.Packet.Header.Sequence
			c.tcb.snd.WL2 = packet.Packet.Header.Ack

		}
	}
	//c.tcb.rcv.NXT += 1
}

func (c *Conn) handleSegment(packet AddressedPacket) error {
	if packet.Packet.Data == nil || len(packet.Packet.Data) == 0 {
		fmt.Println("[info] tcp segment data is empty.")
		return nil
	}
	c.rcvBuffer = append(c.rcvBuffer, packet.Packet.Data...)
	l := len(packet.Packet.Data)
	c.tcb.rcv.NXT = c.tcb.rcv.NXT + uint32(l)
	c.tcb.rcv.WND = c.tcb.rcv.WND - uint32(l)
	if err := c.send(tcp.ACK, nil); err != nil {
		return err
	}
	return nil
}

func (c *Conn) handleFin(packet AddressedPacket) error {
	return c.send(tcp.FIN|tcp.ACK, nil)
}

func (c *Conn) send(flag tcp.ControlFlag, data []byte) error {
	c.tcb.snd.NXT += 1
	c.tcb.rcv.NXT += 1
	p, err := tcp.Build(
		uint16(c.tcb.peer.Port), uint16(c.tcb.peer.PeerPort),
		c.tcb.snd.NXT, c.tcb.rcv.NXT, flag, uint16(c.tcb.rcv.WND), 0, data)
	if err != nil {
		return err
	}
	c.inner.enqueue(c.tcb.peer.PeerAddr, p)
	c.tcb.showSeq()
	return nil
}
