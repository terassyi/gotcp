package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/proto/port"
)

type Conn struct {
	tcb   *controlBlock
	Peer  *port.Peer
	queue chan AddressedPacket
	inner *Tcp
}

func newConn(peer *port.Peer) (*Conn, error) {
	return &Conn{
		tcb:   NewControlBlock(peer),
		Peer:  peer,
		queue: make(chan AddressedPacket, 100),
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

func (c *Conn) Close() error {
	return c.activeClose()
}

func (c *Conn) activeClose() error {
	// close tcb
	if c.tcb.state != ESTABLISHED && c.tcb.state != CLOSE_WAIT && c.tcb.state != SYN_RECVD {
		return fmt.Errorf("invalid state")
	}
	fin, err := tcp.Build(
		uint16(c.tcb.peer.Port), uint16(c.tcb.peer.PeerPort),
		c.tcb.snd.NXT, c.tcb.rcv.NXT,
		tcp.FIN|tcp.ACK,
		0, 0, nil)
	if err != nil {
		return err
	}
	c.inner.enqueue(c.tcb.peer.PeerAddr, fin)
	if c.tcb.state == SYN_RECVD || c.tcb.state == ESTABLISHED {
		fmt.Println("[info] transmission control block state is FIN-WAIT1")
		c.tcb.FIN_WAIT1()
		// wait ack of fin

		fmt.Println("[info] transmission control block state is FIN-WAIT2")
		c.tcb.FIN_WAIT2()
		// wait fin

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
	if c.tcb.rcv.WND == 0 || packet.Packet.Header.Sequence != c.tcb.rcv.NXT {
		if err := c.send(c.tcb.snd.NXT, c.tcb.rcv.NXT, tcp.ACK, nil); err != nil {
			return err
		}
		return fmt.Errorf("recieve window is zero")
	}

	// second check the RST bit,
	if packet.Packet.Header.OffsetControlFlag.ControlFlag().Rst() {
		c.tcb.CLOSED()
		return nil
	}
	// third check security and precedence
	// TODO

	// fourth, check the SYN bit
	if packet.Packet.Header.OffsetControlFlag.ControlFlag().Syn() {
		if err := c.send(c.tcb.snd.NXT, c.tcb.rcv.NXT, tcp.RST, nil); err != nil {
			return err
		}
		c.tcb.CLOSED()
		return nil
	}
	// fifth check the ACK field
	if packet.Packet.Header.OffsetControlFlag.ControlFlag().Ack() {
		switch c.tcb.state {
		case ESTABLISHED:
			c.handleEstablished(packet)
		case FIN_WAIT1:
			c.handleEstablished(packet)
			if c.tcb.finSend {
				c.tcb.FIN_WAIT2()
			}
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
			if err := c.send(c.tcb.snd.NXT, c.tcb.rcv.NXT, tcp.FIN|tcp.ACK, nil); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("ack field is not set")
	}
	// sixth check the URG bit
	// TODO

	// seventh process the segment text

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
}

func (c *Conn) send(seq, ack uint32, flag tcp.ControlFlag, data []byte) error {
	p, err := tcp.Build(
		uint16(c.tcb.peer.Port), uint16(c.tcb.peer.PeerPort),
		seq, ack, flag, uint16(c.tcb.rcv.WND), 0, data)
	if err != nil {
		return err
	}
	c.inner.enqueue(c.tcb.peer.PeerAddr, p)
	return nil
}
