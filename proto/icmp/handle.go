package icmp

import (
	"github.com/terassyi/gotcp/packet/icmp"
	"github.com/terassyi/gotcp/proto"
	"log"
)

type Icmp struct {
	*proto.ProtocolBuffer
}

func New() *Icmp {
	return &Icmp{proto.NewProtocolBuffer()}
}

func (i *Icmp) Recv(buf []byte) {
	i.Buffer <- buf
}

func (i *Icmp) Handle() {
	for {
		buf, ok := <-i.Buffer
		if ok {
			packet, err := icmp.New(buf)
			if err != nil {
				log.Printf("icmp packet serialize error: %v", err)
				return
			}
			packet.Show()
		}
	}
}
