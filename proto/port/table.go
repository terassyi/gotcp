package port

import (
	"fmt"
	"github.com/terassyi/gotcp/packet/ipv4"
	"sync"
)

type Table struct {
	Entry []*Peer
	mutex *sync.RWMutex
}

type Peer struct {
	PeerAddr *ipv4.IPAddress
	PeerPort int
	Port     int
}

func NewPeer(addr *ipv4.IPAddress, peerport, myport int) *Peer {
	return &Peer{
		PeerAddr: addr,
		PeerPort: peerport,
		Port:     myport,
	}
}

const (
	MIN_PORT_RANGE int = 40000
	MAX_PORT_RANGE int = 65535
)

func New() (*Table, error) {
	return &Table{
		Entry: make([]*Peer, 0, 100),
		mutex: &sync.RWMutex{},
	}, nil
}

func (t *Table) Add(addr *ipv4.IPAddress, peerport, srcport int) (*Peer, error) {
	var port int
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if srcport == 0 {
		p, err := t.getAvailablePort(addr, peerport)
		if err != nil {
			return nil, err
		}
		if p == 0 {
			return nil, fmt.Errorf("failed to find available port")
		}
		port = p
	} else {
		port = srcport
	}
	peer := &Peer{
		PeerAddr: addr,
		PeerPort: peerport,
		Port:     port,
	}
	t.Entry = append(t.Entry, peer)
	return peer, nil
}

func (t *Table) Delete(peer *Peer) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	index, ok := t.search(peer)
	if !ok {
		return fmt.Errorf("no such peer")
	}
	if err := t.delete(index); err != nil {
		return err
	}
	return nil
}

func (t *Table) delete(index int) error {
	if index >= len(t.Entry) {
		return fmt.Errorf("invalid index")
	}
	t.Entry = append(t.Entry[:index], t.Entry[index+1:]...)
	return nil
}

func (t *Table) search(peer *Peer) (int, bool) {
	for index, p := range t.Entry {
		if p.PeerAddr == peer.PeerAddr && p.PeerPort == peer.PeerPort && p.Port == peer.Port {
			return index, true
		}
	}
	return -1, false
}

func (t *Table) Search(peer *Peer) (int, bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.search(peer)
}

func (t *Table) searchByPeer(peer *Peer) (int, bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for index, p := range t.Entry {
		if p.PeerAddr == peer.PeerAddr && p.PeerPort == peer.PeerPort {
			return index, true
		}
	}
	return 0, false
}

func (t *Table) searchBySrcPort(srcport int) (int, bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for index, p := range t.Entry {
		if p.Port == srcport {
			return index, true
		}
	}
	return 0, false
}

func (t *Table) getAvailablePort(peeraddr *ipv4.IPAddress, peerport int) (int, error) {
	for port := MIN_PORT_RANGE; port < MAX_PORT_RANGE; port++ {
		used := false
		for _, e := range t.Entry {
			if e.Port == port {
				used = true
				break
			}
		}
		if used {
			continue
		}
		return port, nil
	}
	return 0, nil
}