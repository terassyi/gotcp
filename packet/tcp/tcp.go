package tcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// document http://www5d.biglobe.ne.jp/stssk/rfc/rfc793j.html

type TCPHeader struct {
	SourcePort      uint16 // 16bits
	DestinationPort uint16 // 16bits
	Sequence        uint32 // 32bits
	Ack             uint32 // 32bits
	// Offset          uint8 // 4bits
	// Flag            ControlFlag // 4bits
	OffsetControlFlag OffsetControlFlag // 16bits
	WindowSize        uint16
	Checksum          uint16
	Urgent            uint16
}

type TCPPacket struct {
	Header TCPHeader
	Option []byte
	Data   []byte
}

type OffsetControlFlag uint16

type ControlFlag uint8

// type Option []byte

func newOffsetControlFlag(offset uint8, flag ControlFlag) OffsetControlFlag {
	return OffsetControlFlag(uint16(offset)<<8 | uint16(flag))
}

func (of OffsetControlFlag) Offset() int {
	of8 := uint8(of >> 8)
	return 4 * int(of8>>4)
}

func (of OffsetControlFlag) ControlFlag() ControlFlag {
	return ControlFlag(uint8(of))
}

func (f ControlFlag) String() string {
	var flags []string
	if f.Syn() {
		flags = append(flags, "syn")
	}
	if f.Ack() {
		flags = append(flags, "ack")
	}
	if f.Fin() {
		flags = append(flags, "fin")
	}
	if f.Rst() {
		flags = append(flags, "rst")
	}
	if f.Psh() {
		flags = append(flags, "psh")
	}
	if f.Urg() {
		flags = append(flags, "urg")
	}
	return strings.Join(flags, "|")
}

func (f ControlFlag) Fin() bool {
	if FIN&f != 0 {
		return true
	}
	return false
}

func (f ControlFlag) Syn() bool {
	if SYN&f != 0 {
		return true
	}
	return false
}

func (f ControlFlag) Rst() bool {
	if RST&f != 0 {
		return true
	}
	return false
}

func (f ControlFlag) Psh() bool {
	if PSH&f != 0 {
		return true
	}
	return false
}

func (f ControlFlag) Ack() bool {
	if ACK&f != 0 {
		return true
	}
	return false
}

func (f ControlFlag) Urg() bool {
	if URG&f != 0 {
		return true
	}
	return false
}

func (tcphdr *TCPHeader) PrintTCPHeader() {
	fmt.Println("-------tcp header-------")
	fmt.Printf("source port = %v\n", tcphdr.SourcePort)
	fmt.Printf("destination port = %v\n", tcphdr.DestinationPort)
	fmt.Printf("sequence number = %v\n", tcphdr.Sequence)
	fmt.Printf("ack = %v\n", tcphdr.Ack)
	fmt.Printf("offset = %v\n", tcphdr.OffsetControlFlag.Offset())
	fmt.Printf("control flag = %s\n", tcphdr.OffsetControlFlag.ControlFlag().String())
	fmt.Printf("window size = %v\n", tcphdr.WindowSize)
	fmt.Printf("checksum = %x\n", tcphdr.Checksum)
	fmt.Println("------------------------")
}

func NewTCPPacket(data []byte) (*TCPPacket, error) {
	header := &TCPHeader{}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %v", err)
	}
	optionLength := header.OffsetControlFlag.Offset() - 20
	packet := &TCPPacket{
		Header: *header,
		Option: make([]byte, optionLength),
	}
	if err := binary.Read(buf, binary.BigEndian, packet.Option); err != nil {
		return nil, fmt.Errorf("failed to read option: %v", err)
	}
	packet.Data = data[header.OffsetControlFlag.Offset():]
	// packet.Data = buf.Bytes()
	return packet, nil
}

func (tp *TCPPacket) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	if err := binary.Write(buf, binary.BigEndian, tp.Header); err != nil {
		return nil, fmt.Errorf("failed to write: %v", err)
	}
	if err := binary.Write(buf, binary.BigEndian, tp.Data); err != nil {
		return nil, fmt.Errorf("failed to write: %v", err)
	}
	return buf.Bytes(), nil
}

func BuildTCPPacket(src, dst uint16, seq, ack uint32, flag ControlFlag, window, urgent uint16, data []byte) (*TCPPacket, error) {
	header := &TCPHeader{
		SourcePort:      src,
		DestinationPort: dst,
		Sequence:        seq,
		Ack:             ack,
		WindowSize:      window,
		Checksum:        uint16(0),
		Urgent:          urgent,
	}
	// no option
	cf := newOffsetControlFlag(uint8(20), flag)
	header.OffsetControlFlag = cf
	packet := &TCPPacket{
		Header: *header,
		Data:   data,
	}
	fmt.Println("Build TCP packet >>>>>>>>>>>>>>>>>>>>>>>>>>>")
	packet.Header.PrintTCPHeader()
	return packet, nil
}

func (tcpp *TCPPacket) PrintTCPPacket() {
	tcpp.Header.PrintTCPHeader()
	fmt.Println(string(tcpp.Data))
	fmt.Println("------------------------")
}
