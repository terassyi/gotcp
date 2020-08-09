package ipv4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/terassyi/gotcp/util"
	"strconv"
	"strings"
)

/*
0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |Version|  IHL  | Type of Service|          Total Length         |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |         Identification        |Flags|      Fragment Offset    |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  Time to Live |    Protocol   |         Header Checksum       |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                       Source Address                          |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                    Destination Address                        |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                    Options                    |    Padding    |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

type Header struct {
	VHL      VerIHL              // 8bits
	TOS      uint8               // 8bits
	Length   uint16              // 16bits
	Ident    uint16              // 16bits
	FlOffset FlagsFragmentOffset // 16bits
	TTL      uint8               // 8bits
	Protocol IPProtocol          // 8bits
	Checksum uint16              // 16bits
	Src      IPAddress           // 32bits
	Dst      IPAddress           // 32bits
	// Options  []byte
	// Padding  []byte
}
type VerIHL uint8

func (vi VerIHL) Version() uint8 {
	return uint8(vi) >> 4
}

func (vi VerIHL) IHL() uint8 {
	return uint8(vi) & 0x0F
}

type FlagsFragmentOffset uint16

func (fo FlagsFragmentOffset) Flags() uint8 {
	return uint8(fo) >> 5
}

func (fo FlagsFragmentOffset) FragmentOffset() uint16 {
	return uint16(fo) & 0x1FF
}

type IPAddress [4]byte

func NewIPAddress(addr []byte) IPAddress {
	return IPAddress{addr[0], addr[1], addr[2], addr[3]}
}

func (ipaddr IPAddress) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ipaddr[0], ipaddr[1], ipaddr[2], ipaddr[3])
}

func (ipaddr IPAddress) Bytes() []byte {
	return ipaddr[:]
}

func Address(addr []byte) (*IPAddress, error) {
	if len(addr) != 4 {
		return nil, fmt.Errorf("invalid address %v", addr)
	}
	return &IPAddress{addr[0], addr[1], addr[2], addr[3]}, nil
}

type IPProtocol uint8

type Packet struct {
	Header Header
	// Options   []byte
	// Padding []byte
	OptionPadding []byte
	Data          []byte
}

func (iphdr *Header) Show() {
	fmt.Println("----------ip header----------")
	fmt.Println("version = ", iphdr.VHL.Version())
	fmt.Println("ihl = ", iphdr.VHL.IHL())
	fmt.Printf("tos = %02x\n", iphdr.TOS)
	fmt.Printf("length = %02x\n", iphdr.Length)
	fmt.Printf("identifier = %02x\n", iphdr.Ident)
	fmt.Printf("flags = %02x\n", iphdr.FlOffset.Flags())
	fmt.Printf("fragment offset = %02x\n", iphdr.FlOffset.FragmentOffset())
	fmt.Printf("ttl = %02x\n", iphdr.TTL)
	fmt.Printf("protocol = %s\n", iphdr.Protocol.String())
	fmt.Printf("checksum = %02x\n", iphdr.Checksum)
	fmt.Printf("src = %s\n", iphdr.Src.String())
	fmt.Printf("dst = %s\n", iphdr.Dst.String())
}

func (ipp IPProtocol) String() string {
	switch ipp {
	case IPICMPv4Protocol:
		return "icmp"
	case IPTCPProtocol:
		return "tcp"
	case IPUDPProtocol:
		return "udp"
	default:
		return "(UNKNOWN)"
	}
}

func New(data []byte) (*Packet, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("ip header is too short (%d)", len(data))
	}
	header := &Header{}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, header); err != nil {
		return nil, fmt.Errorf("encoding error: %v", err)
	}
	// header.Show()
	if header.VHL.Version() != uint8(4) {
		return nil, fmt.Errorf("ip version is not ipv4")
	}
	if int(header.VHL.IHL())*4 > len(data) {
		return nil, fmt.Errorf("header length is too short")
	}
	if int(header.Length) > len(data) {
		return nil, fmt.Errorf("header length is too short")
	}
	if int(header.TTL) == 0 {
		return nil, fmt.Errorf("ttl is zero")
	}
	headerLength := int(header.VHL.IHL() << 2)
	// sum := util.Checksum2(data, headerLength, 0)
	// fmt.Printf("checksum [%x]:[%x]\n", sum, header.Checksum)
	// if sum != header.Checksum {
	// return nil, fmt.Errorf("invalid checksum")
	// }
	packet := &Packet{
		Header:        *header,
		OptionPadding: make([]byte, header.VHL.IHL()-20),
	}
	if err := binary.Read(buf, binary.BigEndian, packet.OptionPadding); err != nil {
		return nil, fmt.Errorf("error making option and padding: %v", err)
	}
	// packet.Data = buf.Bytes()
	packet.Data = data[headerLength:int(packet.Header.Length)]
	return packet, nil
}

func (ip *Packet) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	if err := binary.Write(buf, binary.BigEndian, ip.Header); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, ip.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ip *Packet) Handle() {
	ip.Header.Show()
}

func (iphdr *Header) DeclTTL() {
	iphdr.TTL--
}

// func (ip *Packet) CalculateChecksum() bool {

// }

func (ip *Packet) ReCalculateChecksum() error {
	// heavy
	ip.Header.Checksum = uint16(0)
	buf, err := ip.Serialize()
	if err != nil {
		return err
	}
	headerLength := int(ip.Header.VHL.IHL() << 2)
	sum := util.Checksum2(buf, headerLength, 0)
	// fmt.Printf("recalculate checksum:%x\n", sum)
	ip.Header.Checksum = sum
	return nil
}

func Build(src, dst IPAddress, protocol IPProtocol, data []byte) (*Packet, error) {
	//no option
	header := &Header{
		VHL:      VerIHL(0x45),
		TOS:      uint8(0),
		Length:   uint16(20 + len(data)),
		Ident:    uint16(0),
		FlOffset: FlagsFragmentOffset(0),
		TTL:      uint8(8),
		Protocol: protocol,
		Checksum: uint16(0),
		Src:      src,
		Dst:      dst,
	}
	packet := &Packet{
		Header: *header,
		Data:   data,
	}
	if err := packet.ReCalculateChecksum(); err != nil {
		return nil, err
	}
	return packet, nil
}

func StringToIPAddress(addr string) (*IPAddress, error) {
	s := strings.Split(addr, ".")
	var address []byte
	for _, v := range s {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		address = append(address, byte(n))
	}
	return Address(address)
}
