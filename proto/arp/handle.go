package arp

import (
	"fmt"
	"github.com/terassyi/gotcp/ioctl"
	"github.com/terassyi/gotcp/packet/arp"
	"github.com/terassyi/gotcp/packet/ethernet"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/proto"
	"log"
)

// global variable
var ArpTable = NewTable()

type Arp struct {
	*proto.ProtocolBuffer
	Table      *Table
	IpAddress  *ipv4.IPAddress
	MacAddress *ethernet.HardwareAddress
	Updated    chan struct{}
}

func New(table *Table) *Arp {
	return &Arp{
		ProtocolBuffer: proto.NewProtocolBuffer(),
		Table:          table,
		Updated:        make(chan struct{}),
	}
}

func (ap *Arp) SetAddr(name string) error {
	ipaddrByte, err := ioctl.Siocgifaddr(name)
	if err != nil {
		return err
	}
	macaddrByte, err := ioctl.Siocgifhwaddr(name)
	if err != nil {
		return err
	}
	ipaddr, err := ipv4.Address(ipaddrByte)
	if err != nil {
		return err
	}
	macaddr, err := ethernet.Address(macaddrByte)
	if err != nil {
		return err
	}
	ap.MacAddress = macaddr
	ap.IpAddress = ipaddr
	return nil
}

func (ap *Arp) Recv(buf []byte) {
	ap.Buffer <- buf
}

// Handle will be called with goroutine
func (ap *Arp) Handle() {
	for {
		buf, ok := <-ap.Buffer
		if ok {
			packet, err := arp.New(buf)
			if err != nil {
				log.Printf("arp packet serialize error: %v", err)
				return
			}
			//packet.Show()
			if err := ap.manage(packet); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Println("failed to recv")
		}
	}
}

func (ap *Arp) manage(packet *arp.Packet) error {
	switch packet.Header.OpCode {
	case arp.ARP_REPLY:
		return ap.reply(packet)
	case arp.ARP_REQUEST:
		return ap.request(packet)
	default:
	}
	return nil
}

func (ap *Arp) reply(packet *arp.Packet) error {
	macaddr, err := ethernet.Address(packet.SourceHardwareAddress)
	if err != nil {
		return err
	}
	ipaddr, err := ipv4.Address(packet.SourceProtocolAddress)
	if err != nil {
		return err
	}
	e := ap.Table.Search(ipaddr)
	if e == nil {
		if err := ap.Table.Insert(macaddr, ipaddr); err != nil {
			return err
		}
		ap.Table.Show()
		ap.Updated <- struct{}{}
		return nil
	}
	ap.Table.Show()
	ok, err := ap.Table.Update(macaddr, ipaddr)
	if err != nil {
		return err
	}
	ap.Updated <- struct{}{}
	if ok {
		return fmt.Errorf("cannot find an entry")
	}
	return nil
}

func (ap *Arp) request(packet *arp.Packet) error {

	return nil
}

func (ap *Arp) Request(targetProtocolAddress *ipv4.IPAddress) (*arp.Packet, error) {
	header := arp.Header{
		HardwareType: arp.HARDWARE_ETHERNET,
		ProtocolType: arp.PROTOCOL_IPv4,
		HardwareSize: uint8(6),
		ProtocolSize: uint8(4),
		OpCode:       arp.ARP_REQUEST,
	}
	return &arp.Packet{
		Header:                header,
		SourceHardwareAddress: ap.MacAddress.Bytes(),
		SourceProtocolAddress: ap.IpAddress.Bytes(),
		TargetHardwareAddress: ethernet.BroadcastAddress[:],
		TargetProtocolAddress: targetProtocolAddress.Bytes(),
	}, nil
}

func (ap *Arp) Reply(targetHardwareAddress *ethernet.HardwareAddress, targetProtocolAddress *ipv4.IPAddress) (*arp.Packet, error) {
	header := arp.Header{
		HardwareType: arp.HARDWARE_ETHERNET,
		ProtocolType: arp.PROTOCOL_IPv4,
		HardwareSize: uint8(6),
		ProtocolSize: uint8(4),
		OpCode:       arp.ARP_REPLY,
	}
	return &arp.Packet{
		Header:                header,
		SourceHardwareAddress: ap.MacAddress.Bytes(),
		SourceProtocolAddress: ap.IpAddress.Bytes(),
		TargetHardwareAddress: targetHardwareAddress.Bytes(),
		TargetProtocolAddress: targetProtocolAddress.Bytes(),
	}, nil
}
