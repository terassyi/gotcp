package icmp

import (
	"github.com/terassyi/gotcp/pkg/logger"
	"github.com/terassyi/gotcp/pkg/packet/icmp"
	"github.com/terassyi/gotcp/pkg/proto"
)

type Icmp struct {
	*proto.ProtocolBuffer
	logger *logger.Logger
}

func New(debug bool) *Icmp {
	return &Icmp{
		ProtocolBuffer: proto.NewProtocolBuffer(),
		logger:         logger.New(debug, "icmp"),
	}
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
				i.logger.Errorf("icmp packet serialize error: %v", err)
				return
			}
			if i.logger.DebugMode() {
				packet.Show()
			}
		}
	}
}
