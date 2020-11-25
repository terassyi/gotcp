package tcp

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/util"
	"strings"
)

// document http://www5d.biglobe.ne.jp/stssk/rfc/rfc793j.html

type Header struct {
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

type pseudoHeader struct {
	sourceAddress ipv4.IPAddress
	destinationAddress ipv4.IPAddress
	zero uint16
	ptcl uint16
	tcpLength uint32
}

type Packet struct {
	Header Header
	Option Options
	Data   []byte
}

type OffsetControlFlag uint16

type ControlFlag uint8

// type Option []byte

func newOffsetControlFlag(offset uint8, flag ControlFlag) OffsetControlFlag {
	return OffsetControlFlag(uint16(offset/4)<<12 | uint16(flag))
}

func (of OffsetControlFlag) Offset() int {
	of8 := uint8(of >> 8)
	return 4 * int(of8>>4)
}

func (of OffsetControlFlag) changeHeaderLength(diff int) OffsetControlFlag {
	now := of.Offset()
	return newOffsetControlFlag(uint8(now+diff), of.ControlFlag())
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
	if f.Ecn() {
		flags = append(flags, "ecn")
	}
	if f.Cwr() {
		flags = append(flags, "cwr")
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

func (f ControlFlag) Ecn() bool {
	if ECN&f != 0 {
		return true
	}
	return false
}

func (f ControlFlag) Cwr() bool {
	if CWR&f != 0 {
		return true
	}
	return false
}

func (tcphdr *Header) Show() {
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

func New(data []byte) (*Packet, error) {
	header := &Header{}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %v", err)
	}
	optionLength := header.OffsetControlFlag.Offset() - 20
	packet := &Packet{
		Header: *header,
	}
	if optionLength > 0 {
		op := make([]byte, optionLength)
		if err := binary.Read(buf, binary.BigEndian, op); err != nil {
			return nil, fmt.Errorf("failed to read option: %v", err)
		}
		options, err := OptionsFromByte(op)
		if err != nil {
			return nil, err
		}
		packet.Option = options
	}
	packet.Data = data[header.OffsetControlFlag.Offset():]
	// packet.Data = buf.Bytes()
	return packet, nil
}

func (tp *Packet) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	if err := binary.Write(buf, binary.BigEndian, tp.Header); err != nil {
		return nil, fmt.Errorf("failed to write: %v", err)
	}
	if err := binary.Write(buf, binary.BigEndian, tp.Option.Byte()); err != nil {
		return nil, fmt.Errorf("failed to write tcp option: %v", err)
	}
	if err := binary.Write(buf, binary.BigEndian, tp.Data); err != nil {
		return nil, fmt.Errorf("failed to write: %v", err)
	}
	return buf.Bytes(), nil
}

func (tp *Packet) ReCalculateChecksum(src, dst ipv4.IPAddress) error {
	tp.Header.Checksum = uint16(0)
	pseudo, err := newPseudoHeader(src, dst, tp)
	if err != nil {
		return err
	}
	pseudoBytes, err := pseudo.serialize()
	if err != nil {
		return err
	}
	data, err := tp.Serialize()
	if err != nil {
		return err
	}
	buf := append(pseudoBytes, data...)
	sum := util.Checksum2(buf, len(buf), 0)
	tp.Header.Checksum = sum
	return nil
}

func Build(src, dst uint16, seq, ack uint32, flag ControlFlag, window, urgent uint16, data []byte) (*Packet, error) {
	header := &Header{
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
	packet := &Packet{
		Header: *header,
		Data:   data,
	}
	// don't calculate checksum, because of pseudo header
	//if err := packet.ReCalculateChecksum(); err != nil {
	//	return nil, err
	//}
	return packet, nil
}

func (tp *Packet) AddOption(ops Options) {
	totalLength := 0
	for _, op := range ops {
		totalLength += op.Length()
	}
	nopPadding := 4 - (totalLength % 4)
	if nopPadding != 0 {
		for i := 0; i < nopPadding; i++ {
			ops = append(ops, NoOperation{})
		}
	}
	tp.Option = ops
	tp.Header.OffsetControlFlag = tp.Header.OffsetControlFlag.changeHeaderLength(totalLength + nopPadding)
}

func (tp *Packet) Length() uint32 {
	length := 20
	length += len(tp.Option.Byte())
	length += len(tp.Data)
	return uint32(length)
}

func (tp *Packet) Show() {
	tp.Header.Show()
	fmt.Println(hex.Dump(tp.Option.Byte()))
	fmt.Println("------------------------")
	fmt.Println(hex.Dump(tp.Data))
	fmt.Println("------------------------")
}

func newPseudoHeader(src, dst ipv4.IPAddress, packet *Packet) (*pseudoHeader, error) {
	return &pseudoHeader{
		sourceAddress:      src,
		destinationAddress: dst,
		zero:               uint16(0),
		ptcl:               uint16(06),
		tcpLength:          packet.Length(),
	}, nil
}

func (p *pseudoHeader) serialize() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	if err := binary.Write(buf, binary.BigEndian, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}