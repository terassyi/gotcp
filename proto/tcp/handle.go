package tcp

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/tcp"
	"github.com/terassyi/gotcp/proto"
)

type Tcp struct {
	*proto.ProtocolBuffer
	sync chan struct{}
}

func New() (*Tcp, error) {
	return &Tcp{
		ProtocolBuffer: proto.NewProtocolBuffer(),
		sync:           make(chan struct{}),
	}, nil
}

func (t *Tcp) Recv(buf []byte) {
	t.Buffer <- buf
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
		t.sync <- struct{}{}
	}
}

func (t *Tcp) HandlePacket(buf []byte) {

	packet, err := tcp.New(buf)
	if err != nil {
		fmt.Printf("[error] tcp packet serialize error: %v\n", err)
		return
	}
	packet.Show()
}
