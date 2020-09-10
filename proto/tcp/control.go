package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/proto/port"
	"math/rand"
	"sync"
	"time"
)

// reimplementation of https://github.com/pandax381/microps/blob/master/tcp.c
type controlBlock struct {
	peer    *port.Peer
	state   state
	snd     *SendSequence
	rcv     *ReceiveSequence
	retrans chan AddressedPacket
	Window  []byte
	finSend bool
	Mutex   *sync.RWMutex
}

// type CbTable []*ControlBlock

type state int

type SendSequence struct {
	UNA uint32 // send unacknowladged
	NXT uint32 // send next
	WND uint16 // send window
	UP  uint32 // send urgent pointer
	WL1 uint32 // segment sequence number used for last window update
	WL2 uint32 // segment acknowledgement number used for last window update
	ISS uint32 // initial send sequence number
}

type ReceiveSequence struct {
	NXT uint32 // receive next
	WND uint32 // receive window
	UP  uint32 // receive urgent pointer
	IRS uint32 // initial receive sequence number
}

func (s state) String() string {
	switch s {
	case CLOSED:
		return "CLOSED"
	case LISTEN:
		return "LISTEN"
	case SYN_SENT:
		return "SYN_SENT"
	case SYN_RECVD:
		return "SYN_RECVD"
	case ESTABLISHED:
		return "ESTABLISHED"
	case FIN_WAIT1:
		return "FIN_WAIT1"
	case FIN_WAIT2:
		return "FIN_WAIT2"
	case CLOSING:
		return "CLOSING"
	case TIME_WAIT:
		return "TIME_WAIT"
	case CLOSE_WAIT:
		return "CLOSE_WAIT"
	case LAST_ACK:
		return "LAST_ACK"
	default:
		return "UNKNOWN"
	}
}

func (cb *controlBlock) CLOSED() {
	cb.state = CLOSED
}

func (cb *controlBlock) LISTEN() {
	cb.state = LISTEN
}

func (cb *controlBlock) SYN_SENT() {
	cb.state = SYN_SENT
}

func (cb *controlBlock) SYN_RECVD() {
	cb.state = SYN_RECVD
}

func (cb *controlBlock) ESTABLISHED() {
	cb.state = ESTABLISHED
}

func (cb *controlBlock) FIN_WAIT1() {
	cb.state = FIN_WAIT1
}

func (cb *controlBlock) FIN_WAIT2() {
	cb.state = FIN_WAIT2
}

func (cb *controlBlock) CLOSING() {
	cb.state = CLOSING
}

func (cb *controlBlock) TIME_WAIT() {
	cb.state = TIME_WAIT
}

func (cb *controlBlock) CLOSE_WAIT() {
	cb.state = CLOSE_WAIT
}

func (cb *controlBlock) LAST_ACK() {
	cb.state = LAST_ACK
}

func (cb *controlBlock) activeOpen() (*tcp.Packet, error) {
	// client
	// send syn
	// move to SYN_SENT
	if cb.state != CLOSED {
		return nil, fmt.Errorf("invalid state: %v", cb.state.String())
	}
	cb.snd.ISS = Random()
	cb.snd.NXT = cb.snd.ISS + 1
	cb.snd.UNA = cb.snd.ISS
	packet, err := tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), cb.snd.ISS, 0, tcp.SYN, 29200, 0, nil)
	if err != nil {
		return nil, err
	}
	// add option
	t, err := tcp.NewTimeStamp()
	if err != nil {
		return nil, err
	}
	packet.AddOption(tcp.Options{tcp.MaxSegmentSize(1460), tcp.SACKPermitted{}, tcp.WindowScale(7), *t})
	cb.SYN_SENT()
	fmt.Println("[info] transmission control block state is SYN_SENT")
	return packet, nil
}

func (cb *controlBlock) passiveOpen() error {
	// server
	// move to LISTEN
	if cb.state != CLOSED {
		return fmt.Errorf("invalid state: %v", cb.state.String())
	}
	cb.LISTEN()
	fmt.Println("[info] transmission control block state is LISTEN")
	return nil
}

func (cb *controlBlock) HandleEvent(packet *tcp.Packet) (*tcp.Packet, error) {
	// implement based on rfc
	// segment arrives
	switch cb.state {
	//if the State is CLOSED
	case CLOSED:
		if packet.Header.OffsetControlFlag.ControlFlag().Rst() {
			return nil, nil
		}
		if packet.Header.OffsetControlFlag.ControlFlag().Ack() {
			// <SEQ=SEG.ACK><CTL=RST>
			return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), packet.Header.Ack, 0, tcp.RST, windowZero, 0, nil)
		} else {
			// <SEQ=0><ACK=SEG.SEQ+SEG.LEN><CTL=RST,ACK>
			ack := packet.Header.Sequence
			if packet.Header.OffsetControlFlag.ControlFlag().Syn() {
				ack++
			}
			if packet.Data != nil {
				ack += uint32(len(packet.Data))
			}
			if packet.Header.OffsetControlFlag.ControlFlag().Fin() {
				ack++
			}
			return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), 0, ack, tcp.RST, windowZero, 0, nil)
		}
	// if the State is LISTEN
	case LISTEN:
		// first check rst
		if packet.Header.OffsetControlFlag.ControlFlag().Rst() {
			// incoming RST should be ignored
			return nil, nil
		}
		// second check ACK
		if packet.Header.OffsetControlFlag.ControlFlag().Ack() {
			// <SEQ=SEG.ACK><CTL=RST>
			return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), packet.Header.Ack, 0, tcp.RST, windowZero, 0, nil)
		}
		// third check SYN
		if packet.Header.OffsetControlFlag.ControlFlag().Syn() {
			// TODO check security <SEQ=SEG.ACK><CTL=RST>
			// TODO if tcb.PRC < SEG.PRC <SEQ=SEG.ACK><CTL=RST>
			cb.rcv.NXT = packet.Header.Sequence + 1
			cb.rcv.IRS = packet.Header.Sequence
			cb.snd.ISS = Random()
			cb.snd.NXT = cb.snd.ISS + 1
			cb.snd.UNA = cb.snd.ISS
			cb.SYN_RECVD()
			fmt.Println("[info] transmission control block state is SYN_RECVD")
			// <SEQ=ISS><ACK=RCV.NXT><CTL=SYN,ACK>
			return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), cb.snd.ISS, cb.rcv.NXT, tcp.SYN|tcp.ACK, windowZero, 0, nil)
		}
		// fourth other text or control
		return nil, fmt.Errorf("invalid State")
	case SYN_SENT:
		// first check ACK
		if packet.Header.OffsetControlFlag.ControlFlag().Ack() {
			if packet.Header.Ack <= cb.snd.ISS || packet.Header.Ack > cb.snd.NXT {
				// <SEQ=SEG.ACK><CTL=RST>
				if !packet.Header.OffsetControlFlag.ControlFlag().Rst() {
					return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), packet.Header.Ack, 0, tcp.RST, windowZero, 0, nil)
				}
				return nil, fmt.Errorf("discard the segment")
			}
		}
		// second check RST
		if packet.Header.OffsetControlFlag.ControlFlag().Rst() {
			if packet.Header.OffsetControlFlag.ControlFlag().Ack() {
				return nil, fmt.Errorf("connection reset")
			}
			return nil, fmt.Errorf("discar the segment")
		}
		// TODO third check the security and precedence
		// fourth check SYN
		if packet.Header.OffsetControlFlag.ControlFlag().Syn() {
			// This step should be reached only if the ACK is ok, or there is no ACK, and it the segment did not contain a RST.
			cb.rcv.NXT = packet.Header.Sequence + 1
			cb.rcv.IRS = packet.Header.Sequence
			if packet.Header.OffsetControlFlag.ControlFlag().Ack() {
				cb.snd.UNA = packet.Header.Ack
				// TODO delete retransmission queue
				if cb.snd.ISS < cb.snd.UNA {
					// <SEQ=SND.NXT><ACK=RCV.NXT><CTL=ACK>
					cb.state = ESTABLISHED
					return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), cb.snd.NXT, cb.rcv.NXT, tcp.ACK, windowZero, 0, nil)
					// TODO check sixth step
				}
			}
			// <SEQ=ISS><ACK=RCV.NXT><CTL=SYN,ACK>
			cb.state = SYN_RECVD
			return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), cb.snd.ISS, cb.rcv.NXT, tcp.SYN|tcp.ACK, windowZero, 0, nil)
		}
	}
	if packet.Header.Sequence != cb.rcv.NXT {
		return nil, fmt.Errorf("sequence mismatch")
	}
	if packet.Header.OffsetControlFlag.ControlFlag().Syn() || packet.Header.OffsetControlFlag.ControlFlag().Rst() {
		return nil, fmt.Errorf("reset or syn flag are set")
	}
	// fifth check ack
	if !packet.Header.OffsetControlFlag.ControlFlag().Ack() {
		return nil, fmt.Errorf("ack flag is not set")
	}
	// if ack flag is set
	switch cb.state {
	case SYN_RECVD:
		if cb.snd.UNA <= packet.Header.Ack && packet.Header.Ack <= cb.snd.NXT {
			cb.state = ESTABLISHED
			// queue push
		} else {
			// <SEQ=SEG.ACK><CTL=RST>
			return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), packet.Header.Ack, 0, tcp.RST, windowZero, 0, nil)
		}
	}
	switch cb.state {
	case LAST_ACK:
		cb.state = CLOSED
		// pthread_cond_signal
	default:
		// ESTABLISHED, FIN_WAIT1, FIN_WAIT2, CLOSE_WAIT, CLOSING
		if cb.snd.UNA < packet.Header.Ack && packet.Header.Ack <= cb.snd.NXT {
			cb.snd.UNA = packet.Header.Ack
		} else if packet.Header.Ack > cb.snd.NXT {
			return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), cb.snd.NXT, cb.rcv.NXT, tcp.ACK, windowZero, 0, nil)
		}
		// send window update
		if cb.state == FIN_WAIT1 {
			cb.state = FIN_WAIT2
		}
		if cb.state == FIN_WAIT2 {
			// if retransmission queue is empty, tcp can close
		}
		if cb.state == CLOSING {
			if packet.Header.Ack == cb.snd.NXT {
				cb.state = TIME_WAIT
				// pthread_cond_signal
			}
			return nil, nil
		}
	}
	// TODO sixth check URG
	// seventh process the segment text
	if packet.Data != nil {
		switch cb.state {
		case ESTABLISHED, FIN_WAIT1, FIN_WAIT2:
			// handle data
			cb.rcv.NXT = packet.Header.Sequence + uint32(len(packet.Data))
			cb.rcv.WND -= uint32(len(packet.Data))
			cb.Window = append(cb.Window, packet.Data...)
			// cb.window <- packet.Data
			// <SEQ=SND.NXT><ACK=RCV.NXT><CTL=ACK>
			return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), cb.snd.NXT, cb.rcv.NXT, tcp.ACK, windowZero, 0, nil)
			// pthread_cond_signal
		}
	}
	// eighth check FIN
	if packet.Header.OffsetControlFlag.ControlFlag().Fin() {
		cb.rcv.NXT++
		switch cb.state {
		case ESTABLISHED, SYN_RECVD:
			cb.state = CLOSE_WAIT
			// pthread_cond_signal
		case FIN_WAIT1:
			// start time-wait timer and stop other timers
			cb.state = TIME_WAIT
		case FIN_WAIT2:
			// start time-wait timer
			cb.state = TIME_WAIT
		}
		return tcp.Build(uint16(cb.peer.Port), uint16(cb.peer.PeerPort), cb.snd.NXT, cb.rcv.NXT, tcp.ACK, windowZero, 0, nil)
	}
	return nil, fmt.Errorf("not matched any State")
}

func Random() uint32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint32()
}

func (cb *controlBlock) IsReadyRecv() bool {
	switch cb.state {
	case ESTABLISHED:
		return true
	case FIN_WAIT1:
		return true
	case FIN_WAIT2:
		return true
	default:
		return false
	}
}

func (cb *controlBlock) IsReadySend() bool {
	switch cb.state {
	case ESTABLISHED:
		return true
	case CLOSE_WAIT:
		return true
	default:
		return false
	}
}

func (cb *controlBlock) startMSL() {
	time.Sleep(time.Minute)
}

func (cb *controlBlock) showSeq() {
	fmt.Printf("[info] <rcv.nxt=%d snd.nxt=%d\n", cb.rcv.NXT, cb.snd.NXT)
}

//func (cb *ControlBlock) buildWindow() []byte {
//	buf := make([]byte, 65535)
//	cb.Mutex.Lock()
//	defer cb.Mutex.Unlock()
//	num := len(cb.Window)
//	for i := 0; i < num; i++ {
//		b := <-cb.Window
//		buf = append(buf, b...)
//	}
//	return buf
//}

func NewControlBlock(peer *port.Peer) *controlBlock {
	return &controlBlock{
		peer:    peer,
		state:   CLOSED,
		snd:     newSnd(),
		rcv:     newRcv(),
		finSend: false,
		retrans: make(chan AddressedPacket, 100),
		Window:  make([]byte, 0, 65535),
		Mutex:   &sync.RWMutex{},
	}
}

func newSnd() *SendSequence {
	return &SendSequence{
		UNA: 0,
		NXT: 0,
		WND: 0,
		UP:  0,
		WL1: 0,
		WL2: 0,
		ISS: 0,
	}
}

func newRcv() *ReceiveSequence {
	return &ReceiveSequence{
		NXT: 0,
		WND: 1024,
		UP:  0,
		IRS: 0,
	}
}

func (s *SendSequence) Show() {
	fmt.Println("-----send sequence-----")
	fmt.Println("snd.UNA=", s.UNA)
	fmt.Println("snd.NXT=", s.NXT)
	fmt.Println("snd.WND=", s.WND)
	fmt.Println("snd.UP=", s.UP)
	fmt.Println("snd.WL1=", s.WL1)
	fmt.Println("snd.WL2=", s.WL2)
	fmt.Println("snd.ISS=", s.ISS)
}

func (r *ReceiveSequence) Show() {
	fmt.Println("-----recv sequence-----")
	fmt.Println("rcv.NXT=", r.NXT)
	fmt.Println("rcv.WND=", r.WND)
	fmt.Println("rcv.UP=", r.UP)
	fmt.Println("rcv.IRS=", r.IRS)
}
