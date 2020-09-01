package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/proto"
	"github.com/terassyi/gotcp/proto/port"
)

type Tcp struct {
	*proto.ProtocolBuffer
	SendQueue chan AddressedPacket
	SynQueue  chan AddressedPacket
	Table     *port.Table
}

type AddressedPacket struct {
	Packet  *tcp.Packet
	Address *ipv4.IPAddress
}

func New() (*Tcp, error) {
	table, err := port.New()
	if err != nil {
		return nil, err
	}
	return &Tcp{
		ProtocolBuffer: proto.NewProtocolBuffer(),
		SendQueue:      make(chan AddressedPacket, 100),
		Table:          table,
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
			fmt.Println("[error] failed to recv buffer.")
			continue
		}
		fmt.Println("[info] recv tcp packet.")
		packet, err := tcp.New(buf)
		if err != nil {
			fmt.Printf("[error] tcp packet serialize error: %v\n", err)
			continue
		}
		packet.Show()
	}
}

func (t *Tcp) HandlePacket(src *ipv4.IPAddress, buf []byte) {

	packet, err := tcp.New(buf)
	if err != nil {
		fmt.Printf("[error] tcp packet serialize error: %v\n", err)
		return
	}
	packet.Show()

}
