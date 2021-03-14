package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/sirupsen/logrus"
	"github.com/terassyi/gotcp/pkg/proto/icmp/ping"
)

type PingCommand struct {
	Iface string
	Dst   string
	Debug bool
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
	f.BoolVar(&p.Debug, "debug", false, "output debug messages")
}

func (p *PingCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	pin, err := ping.New(p.Iface, p.Dst, p.Debug)
	fmt.Println(p.Dst)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "ping",
		}).Error(err)
		return subcommands.ExitFailure
	}
	if err := pin.Start(); err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "ping",
		}).Error(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
