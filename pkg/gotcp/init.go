package gotcp

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/terassyi/gotcp/pkg/interfaces"
	etherframe "github.com/terassyi/gotcp/pkg/packet/ethernet"
	"github.com/terassyi/gotcp/pkg/proto/arp"
	"github.com/terassyi/gotcp/pkg/proto/ethernet"
	"github.com/terassyi/gotcp/pkg/proto/icmp"
	"github.com/terassyi/gotcp/pkg/proto/ipv4"
	"github.com/terassyi/gotcp/pkg/proto/tcp"
)

const (
	afpacket string = "afpacket"
	tuntap   string = "tuntap"
)

func EthernetInit(debug bool) {

}

func TcpInit(name string, debug bool) (*tcp.Tcp, error) {
	iface, err := interfaces.New(name, afpacket)
	if err != nil {
		return nil, err
	}
	arpProtocol := arp.New(arp.NewTable(), debug)
	if err := arpProtocol.SetAddr(name); err != nil {
		return nil, err
	}
	e, err := ethernet.New(iface, arpProtocol)
	if err != nil {
		return nil, err
	}
	icmpProtocol := icmp.New(debug)
	tcpProtocol, err := tcp.New(debug)
	if err != nil {
		return nil, err
	}
	ip, err := ipv4.New(e, icmpProtocol, tcpProtocol, debug)
	if err != nil {
		return nil, err
	}

	go arpProtocol.Handle()
	go icmpProtocol.Handle()
	go ip.TcpSend()

	rcvQueue := make(chan []byte, 100)
	go func() {
		for {
			buf := make([]byte, 1514)
			_, err := ip.Eth.Recv(buf)
			if err != nil {
				panic(err)
			}
			rcvQueue <- buf
		}
	}()

	go func() {
		for {
			time.Sleep(time.Millisecond * 100) // to work the tcp process, now it has to sleep for goroutine switching maybe...
			buf, ok := <-rcvQueue
			if !ok {
				logrus.WithFields(logrus.Fields{
					"command": "tcp client",
				}).Info("failed to receive from interface.")
			}
			frame, err := etherframe.New(buf)
			if err != nil {
				panic(err)
			}
			switch frame.Type() {
			case etherframe.ETHER_TYPE_IP:
				ip.HandlePacket(frame.Payload())
			case etherframe.ETHER_TYPE_ARP:
				arpProtocol.Recv(frame.Payload())
			case etherframe.ETHER_TYPE_IPV6:
				logrus.WithFields(logrus.Fields{
					"command": "tcp client",
				}).Info("ipv6 is not supported")
			default:
				logrus.WithFields(logrus.Fields{
					"command": "tcp client",
				}).Info("unknown ethernet type.")
			}
		}
	}()
	return ip.Tcp, nil
}
