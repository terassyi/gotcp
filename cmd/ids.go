package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/terassyi/gotcp/ids"
	"github.com/terassyi/gotcp/interfaces"
)

type IdsCommand struct {
	Iface string
}

func (*IdsCommand) Name() string {
	return "ids"
}

func (*IdsCommand) Synopsis() string {
	return "ids"
}

func (*IdsCommand) Usage() string {
	return "gotcp ids -i <interface name>"
}

func (ids *IdsCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&ids.Iface, "i", "", "interface")
}

func (id *IdsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	iface, err := interfaces.New(id.Iface, "afpacket")
	if err != nil {
		panic(err)
	}
	defer iface.Close()

	i := ids.New()
	for {
		buf := make([]byte, 1500)
		_, err := iface.Recv(buf)
		if err != nil {
			panic(err)
		}
		if err := i.Recv(buf); err != nil {
			fmt.Println("")
		}
	}
}
