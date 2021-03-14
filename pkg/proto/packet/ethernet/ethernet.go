package ethernet

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type HardwareAddress [6]byte

type EtherType uint16

type EthernetHeader struct {
	Dst  HardwareAddress
	Src  HardwareAddress
	Type EtherType
}

type EthernetFrame struct {
	Header EthernetHeader
	Data   []byte
}

func (hwaddr HardwareAddress) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", hwaddr[0], hwaddr[1], hwaddr[2], hwaddr[3], hwaddr[4], hwaddr[5])
}

func (hwaddr HardwareAddress) Bytes() []byte {
	return hwaddr[:]
}
func Address(data []byte) (*HardwareAddress, error) {
	addr := &HardwareAddress{}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.LittleEndian, addr); err != nil {
		return nil, err
	}
	return addr, nil
}

func (ethhdr EthernetHeader) Show() {
	fmt.Println("----------ethernet header----------")
	fmt.Printf("dst = %s\n", ethhdr.Dst.String())
	fmt.Printf("src = %s\n", ethhdr.Src.String())
	fmt.Printf("type = %v", ethhdr.Type)
	switch ethhdr.Type {
	case ETHER_TYPE_ARP:
		fmt.Printf("(ARP)\n")
	case ETHER_TYPE_IP:
		fmt.Printf("(IP)\n")
	case ETHER_TYPE_IPV6:
		fmt.Printf("(IPV6)\n")
	default:
		fmt.Printf("(UNKNOWN)\n")
	}
	fmt.Println("----------------------------------")
}

func New(data []byte) (*EthernetFrame, error) {
	frame := &EthernetFrame{}
	header := &EthernetHeader{}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, header); err != nil {
		return nil, err
	}
	frame.Header = *header
	frame.Data = buf.Bytes()
	return frame, nil
}

func (eth *EthernetFrame) Payload() []byte {
	return eth.Data
}

func (eth *EthernetFrame) Type() EtherType {
	return eth.Header.Type
}

func (eth *EthernetFrame) Serialize() ([]byte, error) {
	frame := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(frame, binary.BigEndian, eth.Header)
	if err != nil {
		return nil, err
	}
	err = binary.Write(frame, binary.BigEndian, eth.Data)
	if err != nil {
		return nil, err
	}
	return frame.Bytes(), nil
}

func Build(src, dst HardwareAddress, typ EtherType, data []byte) *EthernetFrame {
	header := EthernetHeader{
		Src:  src,
		Dst:  dst,
		Type: typ,
	}
	return &EthernetFrame{
		Header: header,
		Data:   data,
	}
}
