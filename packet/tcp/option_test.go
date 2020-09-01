package tcp

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestOptionsFromByte(t *testing.T) {
	data := []byte{
		0x02, 0x04, 0x05, 0xb4,
		0x04, 0x02,
		0x08, 0x0a, 0x17, 0x4b, 0x06, 0xf5, 0x00, 0x00, 0x00, 0x00,
		0x01,
		0x03, 0x03, 0x07,
	}
	ops, err := OptionsFromByte(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 5 {
		t.Fatalf("actual length: %d", len(ops))
	}
	n := NoOperation{}
	if n != ops[3] {
		t.Fatalf("actual: %v", ops[3])
	}
}

func TestNewTimeStamp(t *testing.T) {
	ts, err := NewTimeStamp()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(hex.Dump(ts.Byte()))
	if len(ts.Byte()) != ts.Length() {
		t.Fatalf("actual: %d", len(ts.Byte()))
	}
}
