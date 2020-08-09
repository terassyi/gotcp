package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/terassyi/gotcp/interfaces"
	"github.com/terassyi/gotcp/packet/ethernet"
	"github.com/terassyi/gotcp/proto/arp"
	eth "github.com/terassyi/gotcp/proto/ethernet"
	"github.com/terassyi/gotcp/proto/icmp"
	"github.com/terassyi/gotcp/proto/ipv4"
)

type DumpCommand struct {
}

func (d *DumpCommand) Name() string {
	return "dump"
}

func (d *DumpCommand) Synopsis() string {
	return "dump"
}

func (d *DumpCommand) Usage() string {
	return `gotcp dump:
	dump packets received by the interface`
}

func (d *DumpCommand) SetFlags(f *flag.FlagSet) {
	// nop
}

func (d *DumpCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	iface, err := interfaces.New("host1_veth0", "afpacket")
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 512)
	defer iface.Close()

	arpProtocol := arp.New(arp.NewTable())
	e, err := eth.New(iface, arpProtocol)
	icmpProtocol := icmp.New()
	ipv4Protocol, err := ipv4.New(e, icmpProtocol)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	fmt.Println("start to recv")

	go arpProtocol.Handle()
	go ipv4Protocol.Handle()
	go icmpProtocol.Handle()
	for {
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
			ipv4Protocol.Recv(frame.Payload())
		case ethernet.ETHER_TYPE_ARP:
			arpProtocol.Recv(frame.Payload())
		default:
			fmt.Println("unsupported packet type.")
		}
	}
	return subcommands.ExitSuccess
}
