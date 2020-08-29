package tcp

import (
	"encoding/binary"
	"testing"
)

func TestNew(t *testing.T) {
	//data := []byte{
	//	0xdc,0xe4,0x00,0x17,0x8f,0x9d,
	//	0x16,0x21,0x00,0x00,0x00,0x00,0xa0,0x02,
	//	0x72,0x10,0x81,0x84,0x00,0x00,0x02,0x04,
	//	0x05,0xb4,0x04,0x02,0x08,0x0a,0xfd,0x68,
	//	0x8a,0xe4,0x00,0x00,0x00,0x00,0x01,0x03,
	//	0x03,0x07,
	//	}

	data := []byte{0, 23, 220, 234, 0, 0, 0, 0, 202, 42, 25, 13, 80, 20, 0, 0, 110, 65, 0, 0, 2, 4, 5, 180, 4, 2, 8, 10, 253, 191, 26, 95, 0, 0, 0, 0, 1, 3, 3, 7}
	packet, err := New(data)
	if err != nil {
		t.Fatal(err)
	}
	if packet.Header.OffsetControlFlag.ControlFlag() != SYN {
		t.Fatalf("actual: %v", packet.Header.OffsetControlFlag.ControlFlag().String())
	}
}

func TestNewOffsetControlFlag(t *testing.T) {
	oc := newOffsetControlFlag(uint8(40), SYN)
	wanted := binary.BigEndian.Uint16([]byte{0xa0, 0x02})
	if wanted != uint16(oc) {
		t.Fatalf("actual: %b wanted: %b", oc, wanted)
	}
	if oc.Offset() != 40 {
		t.Fatalf("actual: %b wanted: %b", oc.Offset(), 40)
	}
}