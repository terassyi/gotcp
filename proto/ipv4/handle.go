package ipv4

import (
	"fmt"
	"github.com/terassyi/gotcp/logger"
	etherframe "github.com/terassyi/gotcp/packet/ethernet"
	"github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/proto"
	"github.com/terassyi/gotcp/proto/ethernet"
	"github.com/terassyi/gotcp/proto/icmp"
	"github.com/terassyi/gotcp/proto/tcp"
	"syscall"
	"unsafe"
)

type Ipv4 struct {
	*proto.ProtocolBuffer
	Eth     *ethernet.Ethernet
	Address *ipv4.IPAddress
	Icmp    *icmp.Icmp
	Tcp     *tcp.Tcp
	logger  *logger.Logger
}

func New(eth *ethernet.Ethernet, i *icmp.Icmp, tcp *tcp.Tcp, debug bool) (*Ipv4, error) {
	addr, err := siocgifaddr(eth.Name())
	if err != nil {
		return nil, err
	}
	a, err := ipv4.Address(addr)
	if err != nil {
		return nil, err
	}

	return &Ipv4{
		ProtocolBuffer: proto.NewProtocolBuffer(),
		Eth:            eth,
		Address:        a,
		Icmp:           i,
		Tcp:            tcp,
		logger:         logger.New(debug, "ipv4"),
	}, nil
}

func (ip *Ipv4) Show() {
	ip.logger.Info("------ip interface ------")
	ip.logger.Info("name: %v\n", ip.Eth.Name())
	ip.logger.Info("ip addr: %v\n", ip.Address.String())
	ip.logger.Info("mac addr: %v\n", ip.Eth.Address().String())
}

func (ip *Ipv4) Recv(buf []byte) {
	ip.Buffer <- buf
}

func (ip *Ipv4) Handle() {
	for {
		buf, ok := <-ip.Buffer
		if ok {
			packet, err := ipv4.New(buf)
			if err != nil {
				ip.logger.Errorf("ipv4 packet serialize error: %v", err)
				return
			}
			if err := ip.manage(packet); err != nil {
				ip.logger.Error(err)
				return
			}
		}
	}
}

func (ip *Ipv4) HandlePacket(buf []byte) {
	packet, err := ipv4.New(buf)
	if err != nil {
		ip.logger.Errorf("ipv4 packet serialize error: %v", err)
		return
	}
	if err := ip.manage(packet); err != nil {
		ip.logger.Error(err)
		return
	}
}

func (ip *Ipv4) manage(packet *ipv4.Packet) error {

	switch packet.Header.Protocol {
	case ipv4.IPICMPv4Protocol:
		ip.Icmp.Recv(packet.Data)
	case ipv4.IPTCPProtocol:
		//ip.Tcp.Recv(packet.Data)
		go ip.Tcp.HandlePacket(&packet.Header.Src, packet.Data)
	default:
		return fmt.Errorf("unsupported protocol")
	}
	return nil
}

func (ip *Ipv4) Send(dst ipv4.IPAddress, protocol ipv4.IPProtocol, data []byte) (int, error) {
	packet, err := ipv4.Build(*ip.Address, dst, protocol, data)
	if err != nil {
		return 0, err
	}
	if err := packet.ReCalculateChecksum(); err != nil {
		return 0, err
	}
	ipByte, err := packet.Serialize()
	if err != nil {
		return 0, err
	}

	if _, err := ip.Eth.Send(nil, &dst, etherframe.ETHER_TYPE_IP, ipByte); err != nil {
		return 0, err
	}

	return len(ipByte), nil
}

// this function will be called as goroutine
func (ip *Ipv4) TcpSend() {
	for {
		//fmt.Println("[info] waiting tcp packet to send.")
		addrPacket, ok := <-ip.Tcp.SendQueue
		if !ok {
			ip.logger.Error("failed to handle tcp packet for sending")
			continue
		}
		if err := addrPacket.Packet.ReCalculateChecksum(*ip.Address, *addrPacket.Address); err != nil {
			ip.logger.Error("failed to handle tcp packet for sending")
			continue
		}
		data, err := addrPacket.Packet.Serialize()
		if err != nil {
			ip.logger.Error(err)
		}
		_, err = ip.Send(*addrPacket.Address, ipv4.IPTCPProtocol, data)
		if err != nil {
			ip.logger.Error(err)
		}
		//ip.logger.Debugf("[info] %dbytes tcp packet sent\n", l)
	}
}

func siocgifaddr(name string) ([]byte, error) {

	type sockaddr struct {
		family uint16
		addr   [14]byte
	}

	soc, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.Close(soc)
	ifreq := struct {
		name [syscall.IFNAMSIZ]byte
		addr sockaddr
		_pad [8]byte
	}{}
	copy(ifreq.name[:syscall.IFNAMSIZ-1], name)
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(soc), syscall.SIOCGIFADDR, uintptr(unsafe.Pointer(&ifreq))); errno != 0 {
		return nil, errno
	}
	return ifreq.addr.addr[2:6], nil
}
