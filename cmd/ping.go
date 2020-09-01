package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/terassyi/gotcp/proto/icmp/ping"
)

type PingCommand struct {
	Iface string
	Dst   string
}

func (p *PingCommand) Name() string {
	return "ping"
}

func (p *PingCommand) Synopsis() string {
	return "ping"
}

func (p *PingCommand) Usage() string {
	return `goctp ping -i <interface name> -dest <destination address>:
	send icmp echo request packets and receive reply packets`
}

func (p *PingCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.Iface, "i", "", "interface")
	f.StringVar(&p.Dst, "dest", "", "destination address")
	// nop
}

func (p *PingCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	pin, err := ping.New(p.Iface, p.Dst)
	fmt.Println(p.Dst)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	if err := pin.Start(); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
