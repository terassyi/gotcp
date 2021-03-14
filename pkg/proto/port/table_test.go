package port

import (
	"testing"

	"github.com/terassyi/gotcp/pkg/packet/ipv4"
)

func TestAvailablePort(t *testing.T) {
	table, _ := New()
	table.Entry = append(table.Entry, &Peer{
		PeerAddr: &ipv4.IPAddress{192, 168, 0, 2},
		PeerPort: 80,
		Port:     MIN_PORT_RANGE,
	})
	p, _ := table.getAvailablePort(&ipv4.IPAddress{192, 168, 0, 3}, 8080)
	wanted := MIN_PORT_RANGE + 1
	if p != wanted {
		t.Fatalf("failed wanted %d : actual %d", wanted, p)
	}
}

func TestAdd(t *testing.T) {
	table, _ := New()
	table.Entry = append(table.Entry, &Peer{
		PeerAddr: &ipv4.IPAddress{192, 168, 0, 2},
		PeerPort: 80,
		Port:     MIN_PORT_RANGE,
	})
	p, err := table.Add(&ipv4.IPAddress{192, 168, 0, 4}, 8080, 0)
	if err != nil {
		t.Fatal(err)
	}
	if p.Port != MIN_PORT_RANGE+1 {
		t.Fatalf("actual port %d", p.Port)
	}
	if p.PeerPort != 8080 {
		t.Fatalf("actual peer port %d", p.PeerPort)
	}
	if p.PeerAddr.String() != "192.168.0.4" {
		t.Fatalf("actual peer port %v", p.PeerAddr.String())
	}
}

func TestDelete(t *testing.T) {
	table, _ := New()
	peer := &Peer{
		PeerAddr: &ipv4.IPAddress{192, 168, 0, 2},
		PeerPort: 80,
		Port:     MIN_PORT_RANGE,
	}
	table.Entry = append(table.Entry, peer)

	if len(table.Entry) != 1 {
		t.Fatalf("invalid entry length.")
	}
	if err := table.Delete(peer); err != nil {
		t.Fatal(err)
	}
	if len(table.Entry) != 0 {
		t.Fatalf("actutal %d", len(table.Entry))
	}
}
