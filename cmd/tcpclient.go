package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
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
}

func (c *TcpClientCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	iface, err := interfaces.New(c.Iface, "afpacket")
	if err != nil {
		panic(err)
	}

	arpProtocol := arp.New(arp.NewTable())
	if err := arpProtocol.SetAddr(c.Iface); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	e, err := ethernet.New(iface, arpProtocol)
	icmpProtocol := icmp.New()

	tcpProtocol, err := tcp.New()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	ip, err := ipv4.New(e, icmpProtocol, tcpProtocol)
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
				fmt.Println("[info] ipv6 is not supported.")
			default:
				fmt.Println("[info] unknown ethernet type.")
			}
		}
	}()

	// tcp client

	// disable os stack tcp handling
	conn, err := ip.Tcp.Dial(c.Addr, c.Port)
	if err != nil {
		fmt.Printf("[error] %s", err)
		return subcommands.ExitFailure
	}

	fmt.Println(conn)
	time.Sleep(time.Second * 10)
	if err := conn.Close(); err != nil {
		fmt.Printf("[error] %s", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
