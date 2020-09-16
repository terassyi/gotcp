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
	"time"
)

type TcpClientCommand struct {
	Iface string
	Addr  string
	Port  int
	Debug bool
}

func (c *TcpClientCommand) Name() string {
	return "tcpclient"
}

func (c *TcpClientCommand) Synopsis() string {
	return "tcp client"
}

func (c *TcpClientCommand) Usage() string {
	return `gotcp tcpclient -i <interface name> -addr <ip address> -port <port>
	tcp client to destination host`
}

func (c *TcpClientCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.Iface, "i", "", "interface name")
	f.StringVar(&c.Addr, "addr", "", "destination host address")
	f.IntVar(&c.Port, "port", 0, "destination host port")
	f.BoolVar(&c.Debug, "debug", false, "output debug message")
}

func (c *TcpClientCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if c.Debug {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Debug("debug flag is set")
	} else {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Info("debug flag is not set")
	}
	iface, err := interfaces.New(c.Iface, "afpacket")
	if err != nil {
		panic(err)
	}

	arpProtocol := arp.New(arp.NewTable(), c.Debug)
	if err := arpProtocol.SetAddr(c.Iface); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	e, err := ethernet.New(iface, arpProtocol)
	icmpProtocol := icmp.New(c.Debug)

	tcpProtocol, err := tcp.New(c.Debug)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	ip, err := ipv4.New(e, icmpProtocol, tcpProtocol, c.Debug)
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
			buf := make([]byte, 1500)
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
				go ip.HandlePacket(frame.Payload())
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

	// tcp client

	// disable os stack tcp handling
	conn, err := ip.Tcp.Dial(c.Addr, c.Port)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}

	message := "Hello from gotcp client"

	_, err = conn.Write([]byte(message))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}
	logrus.Println("message send> ", message)
	time.Sleep(time.Second * 2)

	buf := make([]byte, 30)
	_, err = conn.Read(buf)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}
	fmt.Println("message recv> ", string(buf))

	if err := conn.Close(); err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
