package tcp

import (
	"sync"

	"github.com/terassyi/gotcp/logger"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/proto"
	"github.com/terassyi/gotcp/proto/port"
)

type Tcp struct {
	*proto.ProtocolBuffer
	SendQueue   chan AddressedPacket
	SynQueue    chan AddressedPacket
	Table       *port.Table
	listeners   map[int]*Listener
	dialers     map[int]*dialer
	connections map[int]*Conn
	mutex       *sync.RWMutex
	logger      *logger.Logger
}

type AddressedPacket struct {
	Packet  *tcp.Packet
	Address *ipv4.IPAddress
}

func New(debug bool) (*Tcp, error) {
	table, err := port.New()
	if err != nil {
		return nil, err
	}
	return &Tcp{
		ProtocolBuffer: proto.NewProtocolBuffer(),
		SendQueue:      make(chan AddressedPacket, 100),
		SynQueue:       make(chan AddressedPacket, 100),
		Table:          table,
		listeners:      make(map[int]*Listener),
		dialers:        make(map[int]*dialer),
		connections:    make(map[int]*Conn),
		mutex:          &sync.RWMutex{},
		logger:         logger.New(debug, "tcp"),
	}, nil
}

func (t *Tcp) Recv(buf []byte) {
	t.Buffer <- buf
}

func (t *Tcp) enqueue(addr *ipv4.IPAddress, packet *tcp.Packet) {
	t.SendQueue <- AddressedPacket{
		Packet:  packet,
		Address: addr,
	}
}

func (t *Tcp) Handle() {
	for {
		buf, ok := <-t.Buffer
		if !ok {
			continue
		}
		packet, err := tcp.New(buf)
		if err != nil {
			t.logger.Errorf("tcp packet serialize error: %v\n", err)
			continue
		}
		packet.Show()
	}
}

func (t *Tcp) HandlePacket(src *ipv4.IPAddress, buf []byte) {

	packet, err := tcp.New(buf)
	if err != nil {
		t.logger.Errorf("tcp packet serialize error: %v\n", err)
		return
	}
	t.logger.Info("---incomming packet in ack initial phase---")
	packet.Header.Show()
	t.logger.Info("-------------------------------------------")

	// handle packet
	// listener
	l, ok := t.listeners[int(packet.Header.DestinationPort)]
	if ok {
		l.queue <- AddressedPacket{
			Packet:  packet,
			Address: src,
		}
		return
	}

	// dialer
	d, ok := t.dialers[int(packet.Header.DestinationPort)]
	if ok {
		d.queue <- AddressedPacket{
			Packet:  packet,
			Address: src,
		}
		return
	}

	// connection
	c, ok := t.connections[int(packet.Header.DestinationPort)]
	if ok {
		if err := c.handle(AddressedPacket{
			Packet:  packet,
			Address: src,
		}); err != nil {
			t.logger.Error(err)
			return
		}
		return
	} else {

	}
	// TODO delete this info.
	t.logger.Info("received packet is not handled. invalid peer.")
	return
}
