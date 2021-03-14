package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/terassyi/gotcp/pkg/interfaces"
	"github.com/terassyi/gotcp/pkg/packet/ethernet"
	"github.com/terassyi/gotcp/pkg/proto/arp"
	eth "github.com/terassyi/gotcp/pkg/proto/ethernet"
	"github.com/terassyi/gotcp/pkg/proto/icmp"
	"github.com/terassyi/gotcp/pkg/proto/ipv4"
	"github.com/terassyi/gotcp/pkg/proto/tcp"
)

type DumpCommand struct {
	Iface string
}

func (d *DumpCommand) Name() string {
	return "dump"
}

func (d *DumpCommand) Synopsis() string {
	return "dump"
}

func (d *DumpCommand) Usage() string {
	return `gotcp dump -i <interface name>:
	dump packets received by the interface`
}

func (d *DumpCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&d.Iface, "i", "", "interface")
}

func (d *DumpCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	iface, err := interfaces.New(d.Iface, "afpacket")
	if err != nil {
		panic(err)
	}
	defer iface.Close()

	arpProtocol := arp.New(arp.NewTable(), false)
	e, err := eth.New(iface, arpProtocol)
	icmpProtocol := icmp.New(false)

	tcpProtocol, err := tcp.New(false)
	if err != nil {
		return subcommands.ExitFailure
	}
	ipv4Protocol, err := ipv4.New(e, icmpProtocol, tcpProtocol, false)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	ipv4Protocol.Show()
	go arpProtocol.Handle()
	//go ipv4Protocol.Handle()
	go icmpProtocol.Handle()

	//go tcpProtocol.Handle()

	for {
		buf := make([]byte, 1500)
		_, err := iface.Recv(buf)
		if err != nil {
			panic(err)
		}
		frame, err := ethernet.New(buf)
		if err != nil {
			panic(err)
		}
		//frame.Header.Show()

		switch frame.Type() {
		case ethernet.ETHER_TYPE_IP:
			//ipv4Protocol.Recv(frame.Payload())
			go ipv4Protocol.HandlePacket(frame.Payload())
		case ethernet.ETHER_TYPE_ARP:
			arpProtocol.Recv(frame.Payload())
		default:
			fmt.Println("unsupported packet type.")
		}
	}
	return subcommands.ExitSuccess
}
