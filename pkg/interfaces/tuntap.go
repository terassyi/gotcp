package interfaces

import (
	"fmt"
	"os"
	"syscall"
)

const tuntap = "/dev/net/tun"

type tunDevice struct {
	file *os.File
	name string
}

func newTunDevice(name string) (*tunDevice, error) {
	name, file, err := openDevice(name)
	if err != nil {
		return nil, err
	}
	return &tunDevice{
		file: file,
		name: name,
	}, nil
}

func (tun *tunDevice) Name() string {
	return tun.name
}

func (tun *tunDevice) Fd() int {
	return int(tun.file.Fd())
}

func (tun *tunDevice) Recv(buf []byte) (int, error) {
	return tun.file.Read(buf)
}

func (tun *tunDevice) Send(buf []byte) (int, error) {
	return tun.file.Write(buf)
}

func (tun *tunDevice) Close() error {
	return tun.file.Close()
}

func (tun *tunDevice) Address() ([]byte, error) {
	addr, err := siocgifhwaddr(tun.name)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func openDevice(name string) (string, *os.File, error) {
	if len(name) >= syscall.IFNAMSIZ {
		return "", nil, fmt.Errorf("name is too long")
	}
	file, err := os.OpenFile(tuntap, os.O_RDWR, 0600)
	if err != nil {
		return "", nil, err
	}
	name, err = tunsetiff(file.Fd(), name, syscall.IFF_TUN|syscall.IFF_NO_PI)
	if err != nil {
		return "", nil, err
	}
	flags, err := siocgifflags(name)
	if err != nil {
		file.Close()
		return "", nil, err
	}
	flags |= (syscall.IFF_UP | syscall.IFF_RUNNING)
	if err := siocsifflags(name, flags); err != nil {
		file.Close()
		return "", nil, err
	}
	return name, file, nil
}
