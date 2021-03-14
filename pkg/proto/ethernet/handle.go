package ethernet

import (
	"fmt"

	"github.com/terassyi/gotcp/pkg/interfaces"
	"github.com/terassyi/gotcp/pkg/packet/ethernet"
	"github.com/terassyi/gotcp/pkg/packet/ipv4"
	"github.com/terassyi/gotcp/pkg/proto"
	"github.com/terassyi/gotcp/pkg/proto/arp"
)

type Ethernet struct {
	*proto.ProtocolBuffer
	iface   interfaces.Iface
	address *ethernet.HardwareAddress
	Arp     *arp.Arp
}

func New(iface interfaces.Iface, arp *arp.Arp) (*Ethernet, error) {
	mac, err := iface.Address()
	if err != nil {
		return nil, err
	}
	addr, err := ethernet.Address(mac)
	if err != nil {
		return nil, err
	}
	return &Ethernet{
		iface:   iface,
		address: addr,
		Arp:     arp,
	}, nil
}

func (e *Ethernet) Name() string {
	return e.iface.Name()
}

func (e *Ethernet) Address() *ethernet.HardwareAddress {
	return e.address
}

func (e *Ethernet) Close() error {
	return e.Close()
}

func (e *Ethernet) Recv(buf []byte) (int, error) {
	return e.iface.Recv(buf)

}

func (e *Ethernet) ipSend(dstmac *ethernet.HardwareAddress, dstip *ipv4.IPAddress, data []byte) (int, error) {
	if dstmac == nil && dstip == nil {
		return 0, fmt.Errorf("dest address is not specified.")
	}
	if dstmac == nil {
		entry := e.Arp.Table.Search(dstip)
		if entry == nil {
			// send Arp request
			req, err := e.Arp.Request(dstip)
			if err != nil {
				return 0, err
			}
			reqByte, err := req.Serialize()
			if err != nil {
				return 0, err
			}
			if err := e.arpSend(&ethernet.BroadcastAddress, reqByte); err != nil {
				return 0, err
			}

			// enqueue in wait queue
			for {
				_, ok := <-e.Arp.Updated
				if ok {
					return e.ipSend(dstmac, dstip, data)
				}
			}
		}
		frame := ethernet.Build(*e.address, *entry.MacAddress, ethernet.ETHER_TYPE_IP, data)
		frameBytes, err := frame.Serialize()
		if err != nil {
			return 0, err
		}
		return e.iface.Send(frameBytes)
	}

	frame := ethernet.Build(*e.address, *dstmac, ethernet.ETHER_TYPE_IP, data)
	frameBytes, err := frame.Serialize()
	if err != nil {
		return 0, err
	}
	return e.iface.Send(frameBytes)
}

func (e *Ethernet) arpSend(dst *ethernet.HardwareAddress, data []byte) error {
	frame := ethernet.Build(*e.address, *dst, ethernet.ETHER_TYPE_ARP, data)
	frameByte, err := frame.Serialize()
	if err != nil {
		return err
	}
	if _, err := e.iface.Send(frameByte); err != nil {
		return err
	}
	return nil
}

func (e *Ethernet) Send(dstmac *ethernet.HardwareAddress, dstip *ipv4.IPAddress, protocol ethernet.EtherType, data []byte) (int, error) {
	switch protocol {
	case ethernet.ETHER_TYPE_ARP:
		return 26, e.arpSend(dstmac, data)
	case ethernet.ETHER_TYPE_IP:
		return e.ipSend(dstmac, dstip, data)
	default:
		return 0, fmt.Errorf("unsupported protocol in ethernet")
	}

}
