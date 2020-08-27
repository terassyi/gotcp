package main

import (
	"context"
	"flag"
	"github.com/google/subcommands"
	"github.com/terassyi/gotcp/cmd"
	"os"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	subcommands.Register(&cmd.DumpCommand{}, "")
	subcommands.Register(&cmd.PingCommand{}, "")
	subcommands.Register(&cmd.TcpClientCommand{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
