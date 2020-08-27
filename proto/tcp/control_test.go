package tcp

import (
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/proto/port"
	"testing"
)

func TestActiveOpen(t *testing.T) {
	peer := port.NewPeer(&ipv4.IPAddress{192, 168, 0, 3}, 8080, 4000)
	cb := NewControlBlock(peer)
	packet, err := cb.activeOpen()
	if err != nil {
		t.Fatal(err)
	}
	if cb.state != SYN_SENT {
		t.Fatalf("actual %s", cb.state.String())
	}
	packet.Show()
}

func TestPassiveOpen(t *testing.T) {
	peer := port.NewPeer(&ipv4.IPAddress{192, 168, 0, 3}, 8080, 4000)
	cb := NewControlBlock(peer)
	err := cb.passiveOpen()
	if err != nil {
		t.Fatal(err)
	}
	if cb.state != LISTEN {
		t.Errorf("actual %s", cb.state.String())
	}
}
