package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/subcommands"
	"github.com/sirupsen/logrus"
	"github.com/terassyi/gotcp/pkg/interfaces"
	etherframe "github.com/terassyi/gotcp/pkg/packet/ethernet"
	"github.com/terassyi/gotcp/pkg/proto/arp"
	"github.com/terassyi/gotcp/pkg/proto/ethernet"
	"github.com/terassyi/gotcp/pkg/proto/icmp"
	"github.com/terassyi/gotcp/pkg/proto/ipv4"
	"github.com/terassyi/gotcp/pkg/proto/tcp"
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

	rcvQueue := make(chan []byte, 100)
	// packet handle
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

	// tcp client

	conn, err := ip.Tcp.Dial(c.Addr, c.Port)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}

	//message := "Hello from gotcp client"

	message := make([]byte, 20480)
	// message := make([]byte, 6000)
	file, err := os.Open("data/random-data")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}
	defer file.Close()
	if _, err := file.Read(message); err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}
	buf := ""
	go func() {
		for {
			b := make([]byte, 1500)
			l, err := conn.Read(b)
			if err != nil {
				if err == io.EOF {
					fmt.Printf("Client > EOF\n")
					break
				}
				panic(err)
			}
			fmt.Printf("Client> Read %v bytes\n", l)
			buf += string(b)
		}
		fmt.Printf("Client> %v\n", buf)
	}()

	_, err = conn.Write([]byte(message))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}
	time.Sleep(time.Second * 5)

	if err := conn.Close(); err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
