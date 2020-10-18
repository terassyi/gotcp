package ids

import (
	"github.com/terassyi/gotcp/logger"
	"github.com/terassyi/gotcp/packet/ethernet"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/packet/tcp"
)

type Ids struct {
	logger *logger.Logger
}

type idsData struct {
	dstIp   *ipv4.IPAddress
	srcIp   *ipv4.IPAddress
	dstPort int
	srcPort int
	proto   string
}

func (i *Ids) show(d *idsData) {
	i.logger.Infof("%v\n", d.dstIp.String(), d.srcIp.String(), d.dstPort, d.srcPort, d.proto)
}

func New() *Ids {
	return &Ids{logger: logger.New(true, "ids")}
}

func (i *Ids) Recv(d []byte) error {
	//fmt.Println(hex.Dump(d))
	frame, err := ethernet.New(d)
	if err != nil {
		return err
	}

	switch frame.Type() {
	case ethernet.ETHER_TYPE_IP:
		ipPacket, err := ipv4.New(frame.Payload())
		if err != nil {
			return err
		}
		data := &idsData{
			dstIp:   &ipPacket.Header.Dst,
			srcIp:   &ipPacket.Header.Src,
			dstPort: 0,
			srcPort: 0,
		}
		switch ipPacket.Header.Protocol {
		case ipv4.IPICMPv4Protocol:
			data.proto = "icmp"
		case ipv4.IPTCPProtocol:
			tcpSegment, err := tcp.New(ipPacket.Data)
			if err != nil {
				return err
			}
			data.dstPort = int(tcpSegment.Header.DestinationPort)
			data.srcPort = int(tcpSegment.Header.SourcePort)
			data.proto = "tcp"
		case ipv4.IPUDPProtocol:
			data.proto = "udp"
		default:
			data.proto = "unsupported"
		}
		i.show(data)
	default:
	}
	return nil
}
