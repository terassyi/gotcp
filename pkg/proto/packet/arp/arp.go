package arp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/terassyi/gotcp/pkg/packet/ethernet"
)

type Header struct {
	HardwareType HardwareType
	ProtocolType ProtocolType
	HardwareSize uint8
	ProtocolSize uint8
	OpCode       OperationCode
}

type Packet struct {
	Header                Header
	SourceHardwareAddress []byte
	SourceProtocolAddress []byte
	TargetHardwareAddress []byte
	TargetProtocolAddress []byte
}

type HardwareType uint16
type ProtocolType uint16
type OperationCode uint16

func (arp *Packet) Show() {
	fmt.Println("---------------arp---------------")
	fmt.Printf("hardware type = %02x\n", arp.Header.HardwareType)
	fmt.Printf("protocol type = %02x\n", arp.Header.ProtocolType)
	fmt.Printf("hardware address size = %02x\n", arp.Header.HardwareSize)
	fmt.Printf("protocol address size = %02x\n", arp.Header.ProtocolSize)
	fmt.Printf("operation code = %s\n", arp.Header.OpCode.String())
	fmt.Printf("src hwaddr = %s\n", printHadrwareAddress(arp.SourceHardwareAddress))
	fmt.Printf("src protoaddr = %s\n", printProtocolAddress(arp.SourceProtocolAddress))
	fmt.Printf("target hwaddr = %s\n", printHadrwareAddress(arp.TargetHardwareAddress))
	fmt.Printf("target protoaddr = %s\n", printProtocolAddress(arp.TargetProtocolAddress))
}

func (op OperationCode) String() string {
	switch op {
	case ARP_REQUEST:
		return "(REQUEST)"
	case ARP_REPLY:
		return "(REPLY)"
	default:
		return "(UNKNOWN)"
	}
}

func New(data []byte) (*Packet, error) {
	arpHeader := &Header{}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, arpHeader); err != nil {
		return nil, err
	}
	arpPacket := &Packet{
		Header:                *arpHeader,
		SourceHardwareAddress: make([]byte, arpHeader.HardwareSize),
		SourceProtocolAddress: make([]byte, arpHeader.ProtocolSize),
		TargetHardwareAddress: make([]byte, arpHeader.HardwareSize),
		TargetProtocolAddress: make([]byte, arpHeader.ProtocolSize),
	}
	if err := binary.Read(buf, binary.BigEndian, arpPacket.SourceHardwareAddress); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, arpPacket.SourceProtocolAddress); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, arpPacket.TargetHardwareAddress); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, arpPacket.TargetProtocolAddress); err != nil {
		return nil, err
	}
	return arpPacket, nil
}

func (arp *Packet) Serialize() ([]byte, error) {
	packet := bytes.NewBuffer(make([]byte, 0))
	if err := binary.Write(packet, binary.BigEndian, arp.Header); err != nil {
		return nil, err
	}
	if err := binary.Write(packet, binary.BigEndian, arp.SourceHardwareAddress); err != nil {
		return nil, err
	}
	if err := binary.Write(packet, binary.BigEndian, arp.SourceProtocolAddress); err != nil {
		return nil, err
	}
	if err := binary.Write(packet, binary.BigEndian, arp.TargetHardwareAddress); err != nil {
		return nil, err
	}
	if err := binary.Write(packet, binary.BigEndian, arp.TargetProtocolAddress); err != nil {
		return nil, err
	}
	return packet.Bytes(), nil
}

func (arp *Packet) Handle() {
	arp.Show()
}

func printHadrwareAddress(hwaddr []byte) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", hwaddr[0], hwaddr[1], hwaddr[2], hwaddr[3], hwaddr[4], hwaddr[5])
}

func printProtocolAddress(addr []byte) string {
	if len(addr) == 4 {
		return fmt.Sprintf("%d.%d.%d.%d", addr[0], addr[1], addr[2], addr[3])
	} else if len(addr) == 16 {
		return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x", addr[0], addr[1], addr[2], addr[3], addr[4], addr[5], addr[6], addr[7], addr[8], addr[9], addr[10], addr[11], addr[12], addr[13], addr[14], addr[15])
	} else {
		return "unknown address"
	}
}

func Request(srcHardwareAddress, srcProtocolAddress, targetProtocolAddress []byte, protocolType ProtocolType) (*Packet, error) {
	var protocolSize uint8
	switch protocolType {
	case PROTOCOL_IPv4:
		protocolSize = uint8(4)
	case PROTOCOL_IPv6:
		protocolSize = uint8(16)
	default:
		return nil, fmt.Errorf("invalid protocol")
	}
	header := Header{
		HardwareType: HARDWARE_ETHERNET,
		ProtocolType: protocolType,
		HardwareSize: uint8(6),
		ProtocolSize: protocolSize,
		OpCode:       ARP_REQUEST,
	}
	return &Packet{
		Header:                header,
		SourceHardwareAddress: srcHardwareAddress,
		SourceProtocolAddress: srcProtocolAddress,
		TargetHardwareAddress: ethernet.BroadcastAddress[:],
		TargetProtocolAddress: targetProtocolAddress,
	}, nil
}

func Reply(srcHardwareAddress, srcProtocolAddress, targetHardwareAddress, targetProtocolAddress []byte, protocolType ProtocolType) (*Packet, error) {
	var protocolSize uint8
	switch protocolType {
	case PROTOCOL_IPv4:
		protocolSize = uint8(4)
	case PROTOCOL_IPv6:
		protocolSize = uint8(16)
	default:
		return nil, fmt.Errorf("invalid protocol")
	}
	header := Header{
		HardwareType: HARDWARE_ETHERNET,
		ProtocolType: protocolType,
		HardwareSize: uint8(6),
		ProtocolSize: protocolSize,
		OpCode:       ARP_REPLY,
	}
	return &Packet{
		Header:                header,
		SourceHardwareAddress: srcHardwareAddress,
		SourceProtocolAddress: srcProtocolAddress,
		TargetHardwareAddress: targetHardwareAddress,
		TargetProtocolAddress: targetProtocolAddress,
	}, nil
}
