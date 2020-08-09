package interfaces

import "fmt"

type Iface interface {
	Name() string
	Fd() int
	Recv([]byte) (int, error)
	Send([]byte) (int, error)
	Close() error
	Address() ([]byte, error)
}

func New(name, typ string) (Iface, error) {
	switch typ {
	case "afpacket":
		return newAfPacket(name)
	case "tun":
		return newTunDevice(name)
	default:
		return nil, fmt.Errorf("invalid type")

	}
}
