package arp

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/terassyi/gotcp/pkg/packet/ethernet"
	"github.com/terassyi/gotcp/pkg/packet/ipv4"
)

type Table struct {
	Entrys []*Entry
	Mutex  sync.RWMutex
}

type Entry struct {
	IpAddress  *ipv4.IPAddress
	MacAddress *ethernet.HardwareAddress
	TimeStamp  time.Time
}

func NewTable() *Table {
	return &Table{
		Entrys: make([]*Entry, 0, 10),
		Mutex:  sync.RWMutex{},
	}
}

func NewEntry(ipaddr *ipv4.IPAddress, macaddr *ethernet.HardwareAddress) *Entry {
	return &Entry{
		IpAddress:  ipaddr,
		MacAddress: macaddr,
		TimeStamp:  time.Now(),
	}
}

func (t *Table) Search(ipaddr *ipv4.IPAddress) *Entry {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	if len(t.Entrys) == 0 || t.Entrys == nil {
		return nil
	}
	for _, e := range t.Entrys {
		if bytes.Equal(e.IpAddress.Bytes(), ipaddr.Bytes()) {
			return e
		}
	}
	return nil
}

func (t *Table) Insert(macaddr *ethernet.HardwareAddress, ipaddr *ipv4.IPAddress) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if len(t.Entrys) > 10 {
		return fmt.Errorf("arp table is full")
	}
	for _, e := range t.Entrys {
		if bytes.Equal(e.IpAddress.Bytes(), ipaddr.Bytes()) {
			return fmt.Errorf("this address pair is already entried")
		}
	}
	e := NewEntry(ipaddr, macaddr)
	t.Entrys = append(t.Entrys, e)
	return nil
}

func (t *Table) Update(macaddr *ethernet.HardwareAddress, ipaddr *ipv4.IPAddress) (bool, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	for _, e := range t.Entrys {
		if bytes.Equal(e.IpAddress.Bytes(), ipaddr.Bytes()) {
			e.MacAddress = macaddr
			e.TimeStamp = time.Now()
			return true, nil
		}
	}
	return false, nil
}

func (t *Table) Show() {
	fmt.Println("---------------arp table---------------")
	for _, e := range t.Entrys {
		fmt.Printf("hwaddr= %s\n", e.MacAddress.String())
		fmt.Printf("protoaddr=%s\n", e.IpAddress.String())
		fmt.Printf("time=%v\n", e.TimeStamp)
	}
	fmt.Println("---------------------------------------")
}
