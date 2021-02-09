package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/logger"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/proto/port"
	"time"
)

type Conn struct {
	tcb        *controlBlock
	Peer       *port.Peer
	retransmissionQueue      chan *AddressedPacket
	closeQueue chan AddressedPacket
	receivedAck chan uint32
	rcvBuffer  []byte
	readyQueue chan []byte
	inner      *Tcp
	pushFlag   bool
	logger     *logger.Logger
}

const (
	window int = 1024
	rto int = 30 // select a better value
	)

func newConn(peer *port.Peer, debug bool) (*Conn, error) {
	conn := &Conn{
		tcb:        NewControlBlock(peer, debug),
		Peer:       peer,
		retransmissionQueue:      make(chan *AddressedPacket, 100),
		rcvBuffer:  make([]byte, 0, window),
		readyQueue: make(chan []byte, 10),
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
	c.tcb.mutex.RLock()
	defer c.tcb.mutex.RUnlock()
	if c.tcb.state != ESTABLISHED && c.tcb.state != CLOSE_WAIT && c.tcb.state != SYN_RECVD {
		return fmt.Errorf("invalid state")
	}
	if err := c.send(tcp.ACK|tcp.FIN, nil); err != nil {
		return err
	}
	c.tcb.finSend = true
	if c.tcb.state == SYN_RECVD || c.tcb.state == ESTABLISHED {
		c.tcb.FIN_WAIT1()
		// wait ack of fin
		p, ok := <-c.closeQueue
		if !ok {
			return fmt.Errorf("failed to recieve ack of fin.")
		}
		c.tcb.rcv.NXT -= 1
		if p.Packet.Header.OffsetControlFlag.ControlFlag().Fin() {
			// simultaneous close
			c.tcb.rcv.NXT += 1
			if err := c.send(tcp.ACK, nil); err != nil {
				return err
			}
			c.tcb.CLOSING()
			c.tcb.TIME_WAIT()
			c.logger.Debug("start timer")
			c.tcb.startMSL()
			c.tcb.CLOSED()
			return nil
		}
		// got ack

		c.tcb.FIN_WAIT2()
		// wait fin
		_, ok = <-c.closeQueue
		if !ok {
			return fmt.Errorf("failed to recieve ack of fin.")
		}

		c.tcb.TIME_WAIT()
		if err := c.send(tcp.ACK, nil); err != nil {
			return err
		}
		c.tcb.startMSL()
		c.tcb.CLOSED()
		delete(c.inner.connections, c.Peer.Port)
		c.logger.Info("connection closed.")

	} else {
		c.tcb.CLOSE_WAIT()
	}
	return nil
}

func (c *Conn) passiveClose(fin AddressedPacket) error {
	// close tcb
	c.tcb.mutex.RLock()
	defer c.tcb.mutex.RUnlock()
	c.tcb.rcv.NXT += 1
	if c.tcb.state == CLOSED || c.tcb.state == LISTEN || c.tcb.state == SYN_SENT {
		// drop packet
		return nil
	}
	if c.tcb.state == SYN_RECVD || c.tcb.state == ESTABLISHED {
		c.tcb.CLOSE_WAIT()
		if err := c.send(tcp.ACK, nil); err != nil {
			return err
		}
		c.tcb.snd.NXT -= 1
		c.tcb.LAST_ACK()
		if err := c.send(tcp.ACK|tcp.FIN, nil); err != nil {
			return err
		}
		c.tcb.finSend = true
	}
	if c.tcb.state == FIN_WAIT1 {
		if c.tcb.finSend {
			c.tcb.TIME_WAIT()
			// start 2MSL
			// delete TCB
		} else {
			c.tcb.CLOSING()
		}
	}
	if c.tcb.state == FIN_WAIT2 {
		c.tcb.TIME_WAIT()
	}
	if c.tcb.state == CLOSING {
		c.tcb.TIME_WAIT()
	}
	if c.tcb.state == LAST_ACK {
		c.tcb.CLOSED()
	}
	if c.tcb.state == TIME_WAIT {
		// restart 2MSL
		c.tcb.startMSL()
		c.tcb.startMSL()
	}
	delete(c.inner.connections, c.Peer.Port)
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
		//if err := c.send(tcp.ACK, nil); err != nil {
		//	return err
		//}
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
		c.tcb.rcv.NXT += 1
		if err := c.send(tcp.RST, nil); err != nil {
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
			c.tcb.rcv.NXT -= 1
		case FIN_WAIT1:
			c.handleEstablished(packet)
			if c.tcb.finSend {
				fmt.Println("sent to close channel")
				c.closeQueue <- packet
			}
			//fmt.Println("sent to close channel")
			//c.closeQueue <- packet
			return nil
		case FIN_WAIT2:
			c.handleEstablished(packet)
			c.tcb.rcv.NXT -= 1
			//if len(c.tcb.retrans) == 0 {
			//	c.tcb.TIME_WAIT()
			//}
			// ack of fin
			if packet.Packet.Data == nil {
				c.closeQueue <- packet
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
			c.logger.Debug("resend fin")
			if err := c.send(tcp.FIN|tcp.ACK, nil); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("ack field is not set")
	}
	// sixth check the URG bit
	// TODO

	// seventh process the segment text
	switch c.tcb.state {
	case ESTABLISHED, FIN_WAIT1, FIN_WAIT2:
		if err := c.handleSegment(packet); err != nil {
			return err
		}
	default:
		// ignore the segment
	}
	// eighth check fin bit
	if packet.Packet.Header.OffsetControlFlag.ControlFlag().Fin() {
		switch c.tcb.state {
		case CLOSED, LISTEN, SYN_SENT:
			// ignore
		case SYN_RECVD, ESTABLISHED:
			if err := c.passiveClose(packet); err != nil {
				return err
			}
		case FIN_WAIT1:
			// ack
		case FIN_WAIT2:
			c.closeQueue <- packet
			//if err := c.handleFin(packet); err != nil {
			if err := c.passiveClose(packet); err != nil {
				return err
			}
			c.tcb.TIME_WAIT()
		default:
			// stay
		}
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
	c.tcb.rcv.NXT += 1
	// send signal retransmission routine
	c.receivedAck <- packet.Packet.Header.Ack
	c.logger.Debug("sent ack number to retransmission routine")
}

func (c *Conn) handleSegment(packet AddressedPacket) error {
	if packet.Packet.Data == nil || len(packet.Packet.Data) == 0 {
		return nil
	}

	// check PSH
	if packet.Packet.Header.OffsetControlFlag.ControlFlag().Psh() {
		c.readyQueue <- packet.Packet.Data
	} else {
		c.rcvBuffer = append(c.rcvBuffer, packet.Packet.Data...)
		if len(c.rcvBuffer) >= cap(c.rcvBuffer) {
			c.readyQueue <- c.rcvBuffer
			c.rcvBuffer = make([]byte, 0, window)
		}
	}
	l := len(packet.Packet.Data)
	c.tcb.rcv.NXT = c.tcb.rcv.NXT + uint32(l)
	c.tcb.rcv.WND = c.tcb.rcv.WND - uint32(l)

	if err := c.send(tcp.ACK, nil); err != nil {
		return err
	}
	c.tcb.snd.NXT -= 1
	return nil
}

func (c *Conn) handleFin(packet AddressedPacket) error {
	return c.send(tcp.ACK, nil)
}

func (c *Conn) send(flag tcp.ControlFlag, data []byte) error {
	p, err := tcp.Build(
		uint16(c.tcb.peer.Port), uint16(c.tcb.peer.PeerPort),
		c.tcb.snd.NXT, c.tcb.rcv.NXT, flag, uint16(c.tcb.rcv.WND), 0, data)
	if err != nil {
		return err
	}
	c.inner.enqueue(c.tcb.peer.PeerAddr, p)
	if data == nil {
		c.tcb.snd.NXT += 1
	} else {
		c.tcb.snd.NXT += uint32(len(data))
	}
	// add retransmission queue
	if c.tcb.IsReadyRecv() && data != nil {
		c.retransmissionQueue <- &AddressedPacket{
			Packet:  p,
			Address: c.tcb.peer.PeerAddr,
		}
	}
	//c.tcb.showSeq()
	return nil
}

func (c *Conn) resend(packet *AddressedPacket) error {
	c.inner.enqueue(c.tcb.peer.PeerAddr, packet.Packet)
	return nil
}

func (c *Conn) Read(b []byte) (int, error) {
	if !c.tcb.IsReadyRecv() {
		return 0, fmt.Errorf("invalid state")
	}
	return c.read(b)
}

func (c *Conn) read(b []byte) (int, error) {
	buf, ok := <-c.readyQueue
	if !ok {
		return 0, fmt.Errorf("failed to read")
	}
	l := copy(b, buf)
	return l, nil
}

func (c *Conn) Write(b []byte) (int, error) {
	if !c.tcb.IsReadySend() {
		return 0, fmt.Errorf("invalid state")
	}
	return c.write(b)
}

func (c *Conn) write(b []byte) (int, error) {
	flag := tcp.ACK
	if c.pushFlag {
		flag += tcp.PSH
	}
	if err := c.send(flag, b); err != nil {
		return 0, err
	}
	// TODO retransmission handle
	return len(b), nil
}

func (c *Conn) retransmissionHandler() {

	queue := make([]retransmissionPacket, 0, 100)

	go func() {
		ticker := time.NewTicker(1 * time.Second) // scheduler
		for {
			select {
			case q :=  <- c.retransmissionQueue:
				queue = append(queue, retransmissionPacket{
					timeStamp: time.Now(),
					ackNum:    q.Packet.Header.Sequence + uint32(len(q.Packet.Data)),
					packet:    q,
				})
				//fmt.Printf("[DEBUG] append retransmission queue ack=%x data=%s length=%d\n", q.Packet.Header.Sequence, string(q.Packet.Data), len(queue))
			case ack := <- c.receivedAck:
				//fmt.Printf("[DEUBG] ack(=%x) is detected\n", ack)
				idx := -1
				for i, p := range queue {
					if p.ackNum == ack {
						idx = i
						//fmt.Printf("[DEBUG] expected ack=%x\n", ack)
					}
				}
				if idx == -1 {
					continue
				}
				if idx == 0 {
					queue = queue[1:]
				} else if idx == len(queue) {
					queue = queue[:len(queue)-1]
				} else {
					queue = append(queue[:idx-1], queue[idx:]...)
				}
				//fmt.Printf("[DEBUG] retrans queue length is %d\n", len(queue))
			case n := <- ticker.C:
				for _, q := range queue {
					if n.Unix() >= q.timeStamp.Unix() + int64(rto) {
						if err := c.resend(q.packet); err != nil {
							c.logger.Error(err)
							continue
						}
						//fmt.Printf("[DEBUG] timeout retransmit ack=%x\n", q.ackNum)
						q.timeStamp = n // reset timestamp
					}
				}
			}
		}
	}()
}