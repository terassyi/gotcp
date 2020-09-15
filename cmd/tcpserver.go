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
)

type TcpServerCommand struct {
	Iface string
	Port  int
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
}

func (s *TcpServerCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	iface, err := interfaces.New(s.Iface, "afpacket")
	if err != nil {
		panic(err)
	}

	arpProtocol := arp.New(arp.NewTable())
	if err := arpProtocol.SetAddr(s.Iface); err != nil {
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

	// tcp server
	fmt.Println("[info] tcp server prepare... port=", s.Port)
	listener, err := tcpProtocol.Listen("0.0.0.0", s.Port)
	if err != nil {
		fmt.Printf("[error] %v", err)
		return subcommands.ExitFailure
	}
	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("[error] %v", err)
		return subcommands.ExitFailure
	}

	var errChan chan error

	go func() {
		buf := make([]byte, 30)
		l, err := conn.Read(buf)
		if err != nil {
			fmt.Println("[error] ", err)
			errChan <- err
			return
		}
		fmt.Println("[info] message recv: ", string(buf))

		message := "Hello from gotcp server"
		l, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("[error] ", err)
			errChan <- err
			return
		}
		fmt.Printf("[info] message send %dbytes\n", l)

	}()

	err = <-errChan
	if err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
