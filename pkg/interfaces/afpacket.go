package interfaces

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

type afPacket struct {
	fd   int
	name string
}

func newAfPacket(name string) (*afPacket, error) {
	fd, err := openPFPacket(name)
	if err != nil {
		return nil, err
	}
	return &afPacket{
		fd:   fd,
		name: name,
	}, nil
}

func (af *afPacket) Name() string {
	return af.name
}

func (af *afPacket) Fd() int {
	return af.fd
}

func (af *afPacket) Recv(buf []byte) (int, error) {
	return syscall.Read(af.fd, buf)
}

func (af *afPacket) Send(buf []byte) (int, error) {
	return syscall.Write(af.fd, buf)
}

func (af *afPacket) Close() error {
	return syscall.Close(af.fd)
}

func (af *afPacket) Address() ([]byte, error) {
	addr, err := siocgifhwaddr(af.name)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func openPFPacket(name string) (int, error) {
	if name == "" {
		return -1, fmt.Errorf("name is empty")
	}
	if len(name) >= syscall.IFNAMSIZ {
		return -1, fmt.Errorf("name is too long")
	}
	protocol := hton16(syscall.ETH_P_ALL)
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(protocol))
	if err != nil {
		return -1, fmt.Errorf("socket open error %v", err)
	}
	index, err := siocgifindex(name)
	if err != nil {
		syscall.Close(fd)
		return -1, fmt.Errorf("siogifindex error: %v", err)
	}
	addr := &syscall.SockaddrLinklayer{
		Protocol: protocol,
		Ifindex:  int(index),
	}
	if err = syscall.Bind(fd, addr); err != nil {
		syscall.Close(fd)
		return -1, err
	}
	flags, err := siocgifflags(name)
	if err != nil {
		syscall.Close(fd)
		return -1, fmt.Errorf("siocgifflags error: %v", err)
	}
	flags |= syscall.IFF_PROMISC
	if err := siocsifflags(name, flags); err != nil {
		syscall.Close(fd)
		return -1, fmt.Errorf("siocsifflags error: %v", err)
	}
	return fd, nil
}

func hton16(i uint16) uint16 {
	var ret uint16
	binary.BigEndian.PutUint16((*[2]byte)(unsafe.Pointer(&ret))[:], i)
	return ret
}
