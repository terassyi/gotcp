package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/sirupsen/logrus"
	"github.com/terassyi/gotcp/interfaces"
	etherframe "github.com/terassyi/gotcp/packet/ethernet"
	"github.com/terassyi/gotcp/proto/arp"
	"github.com/terassyi/gotcp/proto/ethernet"
	"github.com/terassyi/gotcp/proto/icmp"
	"github.com/terassyi/gotcp/proto/ipv4"
	"github.com/terassyi/gotcp/proto/tcp"
)

type TcpServerCommand struct {
	Iface string
	Port  int
	Debug bool
}

func (*TcpServerCommand) Name() string {
	return "tcpserver"
}

func (*TcpServerCommand) Synopsis() string {
	return "tcp server"
}

func (*TcpServerCommand) Usage() string {
	return `gotcp tcpserver -i <interface name> -port <port>
	tcp server binding port`
}

func (s *TcpServerCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.Iface, "i", "", "interface")
	f.IntVar(&s.Port, "port", 0, "binding port")
	f.BoolVar(&s.Debug, "debug", false, "output debug message")
}

func (s *TcpServerCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if s.Debug {
		logrus.WithFields(
			logrus.Fields{
				"command": "tcp server",
			}).Debug("debug flag is set")
	} else {
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Info("debug flag is not set")
	}
	iface, err := interfaces.New(s.Iface, "afpacket")
	if err != nil {
		panic(err)
	}

	arpProtocol := arp.New(arp.NewTable(), s.Debug)
	if err := arpProtocol.SetAddr(s.Iface); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	e, err := ethernet.New(iface, arpProtocol)
	icmpProtocol := icmp.New(s.Debug)

	tcpProtocol, err := tcp.New(s.Debug)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	ip, err := ipv4.New(e, icmpProtocol, tcpProtocol, s.Debug)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	//defer ip.Eth.Close()

	go arpProtocol.Handle()
	go icmpProtocol.Handle()

	go ip.TcpSend()

	// packet handle
	go func() {
		for {
			buf := make([]byte, 1514)
			_, err := ip.Eth.Recv(buf)
			if err != nil {
				panic(err)
			}
			frame, err := etherframe.New(buf)
			if err != nil {
				panic(err)
			}
			switch frame.Type() {
			case etherframe.ETHER_TYPE_IP:
				//go ip.HandlePacket(frame.Payload())
				ip.HandlePacket(frame.Payload())
			case etherframe.ETHER_TYPE_ARP:
				arpProtocol.Recv(frame.Payload())
			case etherframe.ETHER_TYPE_IPV6:
				logrus.WithFields(logrus.Fields{
					"command": "tcp server",
				}).Info("ipv6 is not supported")
			default:
				logrus.WithFields(logrus.Fields{
					"command": "tcp server",
				}).Info("unknown ethernet type.")
			}
		}
	}()

	// tcp server
	logrus.WithFields(logrus.Fields{
		"command": "tcp client",
	}).Infof("tcp server running at %d\n", s.Port)
	listener, err := tcpProtocol.Listen("0.0.0.0", s.Port)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Error(err)
		return subcommands.ExitFailure
	}
	conn, err := listener.Accept()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Error(err)
		return subcommands.ExitFailure
	}

	var errChan chan error

	go func() {
		buf := make([]byte, 20480)
		l, err := conn.Read(buf)
		if err != nil {
			errChan <- err
			return
		}
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Infof("message recv %d bytes\n", l)
		logrus.Println("message recv> ", string(buf))
		message := "Hello from gotcp server"
		l, err = conn.Write([]byte(message))
		if err != nil {
			errChan <- err
			return
		}
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Infof("message send %dbytes\n", l)
		logrus.Println("message send> ", message)
	}()

	err = <-errChan
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Error(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
