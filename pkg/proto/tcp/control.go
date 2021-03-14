package tcp

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/terassyi/gotcp/pkg/logger"
	"github.com/terassyi/gotcp/pkg/packet/tcp"
	"github.com/terassyi/gotcp/pkg/proto/port"
)

type controlBlock struct {
	peer    *port.Peer
	state   state
	snd     *SendSequence
	rcv     *ReceiveSequence
	retrans chan AddressedPacket
	ack     chan uint32
	Window  []byte
	finSend bool
	mutex   *sync.RWMutex
	logger  *logger.Logger
}

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
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) LISTEN() {
	cb.state = LISTEN
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) SYN_SENT() {
	cb.state = SYN_SENT
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) SYN_RECVD() {
	cb.state = SYN_RECVD
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) ESTABLISHED() {
	cb.state = ESTABLISHED
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) FIN_WAIT1() {
	cb.state = FIN_WAIT1
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) FIN_WAIT2() {
	cb.state = FIN_WAIT2
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) CLOSING() {
	cb.state = CLOSING
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) TIME_WAIT() {
	cb.state = TIME_WAIT
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) CLOSE_WAIT() {
	cb.state = CLOSE_WAIT
	cb.logger.Debug(cb.state.String())
}

func (cb *controlBlock) LAST_ACK() {
	cb.state = LAST_ACK
	cb.logger.Debug(cb.state.String())
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
	packet.AddOption(tcp.Options{tcp.MaxSegmentSize(1460), tcp.WindowScale(7), *t})
	cb.SYN_SENT()
	return packet, nil
}

func (cb *controlBlock) passiveOpen() error {
	// server
	// move to LISTEN
	if cb.state != CLOSED {
		return fmt.Errorf("invalid state: %v", cb.state.String())
	}
	cb.LISTEN()
	return nil
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
	time.Sleep(time.Second * 10)
}

func (cb *controlBlock) showSeq() string {
	return fmt.Sprintf("<rcv.nxt=%d snd.nxt=%d>\n", cb.rcv.NXT, cb.snd.NXT)
}

func (cb *controlBlock) block() error {
	return nil
}

func NewControlBlock(peer *port.Peer, debug bool) *controlBlock {
	return &controlBlock{
		peer:    peer,
		state:   CLOSED,
		snd:     newSnd(),
		rcv:     newRcv(),
		finSend: false,
		retrans: make(chan AddressedPacket, 100),
		Window:  make([]byte, 0, 65535),
		mutex:   &sync.RWMutex{},
		logger:  logger.New(debug, "tcp"),
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
