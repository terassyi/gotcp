package icmp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/terassyi/gotcp/pkg/util"
)

type Header struct {
	Type Type
	// Code messageCode
	Code     uint8
	Checksum uint16
}

type Packet struct {
	Header Header
	Data   []byte
}

type Type uint8
type messageCode uint8

type EchoMessage struct {
	Ident uint16
	Seq   uint16
	Data  []byte
}

func newHeader(typ Type, code uint8) *Header {
	return &Header{
		Type:     typ,
		Code:     code,
		Checksum: uint16(0),
	}
}

func New(data []byte) (*Packet, error) {
	header := &Header{}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, header); err != nil {
		return nil, fmt.Errorf("encoding error: %v", err)
	}
	packet := &Packet{
		Header: *header,
		Data:   buf.Bytes(),
	}
	if err := packet.RecalculateChecksum(); err != nil {
		return nil, err
	}
	return packet, nil
}

func (icmp *Packet) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	if err := binary.Write(buf, binary.BigEndian, icmp.Header); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, icmp.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (icmphdr *Header) Show() {
	fmt.Println("----------icmp header----------")
	fmt.Printf("type = %v(%d)\n", icmphdr.Type, icmphdr.Type)
	fmt.Printf("code = %02x\n", icmphdr.Code)
	fmt.Printf("checksum = %02x\n", icmphdr.Checksum)
}

func (icmp Packet) Show() {
	icmp.Header.Show()
	fmt.Printf("data = %v\n", icmp.Data)
}

func (icmp *Packet) RecalculateChecksum() error {
	icmp.Header.Checksum = uint16(0)
	buf, err := icmp.Serialize()
	if err != nil {
		return err
	}
	sum := util.Checksum2(buf, len(buf), 0)
	icmp.Header.Checksum = sum
	return nil
}

func (typ Type) String() string {
	switch typ {
	case Echo:
		return fmt.Sprintf("Echo Request")
	case EchoReply:
		return fmt.Sprintf("Echo Reply")
	case DestinationUnreachable:
		return fmt.Sprintf("Destination Unreachable")
	case TimeExceeded:
		return fmt.Sprintf("Time Exceeded")
	case Redirect:
		return fmt.Sprintf("Redirect")
	default:
		return fmt.Sprintf("unknown")
	}
}

func (icmp *Packet) Handle() (*Packet, error) {
	switch icmp.Header.Type {
	case Echo:
		return Build(EchoReply, EchoReplyCode, nil)
	case EchoReply:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupportted ICMP type: %v", icmp.Header.Type.String())
	}
}

func Build(typ Type, code uint8, data []byte) (*Packet, error) {
	header := newHeader(typ, code)
	packet := &Packet{
		Header: *header,
		Data:   data,
	}
	buf, err := packet.Serialize()
	if err != nil {
		return nil, err
	}
	sum := util.Checksum2(buf, len(buf), 0)
	packet.Header.Checksum = sum
	return packet, nil
}

func NewEchoMessage(data []byte) (*EchoMessage, error) {
	message := &EchoMessage{}
	message.Ident = binary.BigEndian.Uint16(data[0:2])
	message.Seq = binary.BigEndian.Uint16(data[2:4])
	message.Data = data[4:]
	return message, nil
}

func (e *EchoMessage) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 4))
	if err := binary.Write(buf, binary.BigEndian, e.Ident); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, e.Seq); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, e.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
