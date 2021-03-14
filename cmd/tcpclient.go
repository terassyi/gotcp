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
	"github.com/terassyi/gotcp/pkg/gotcp"
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
	// // tcp client
	gt, err := gotcp.TcpInit(c.Iface, c.Debug)
	conn, err := gt.Dial(c.Addr, c.Port)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}

	message := make([]byte, 20480)
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
	fmt.Printf("Client> Write %d bytes\n", len(message))
	time.Sleep(time.Second * 5)

	if err := conn.Close(); err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp client",
		}).Error(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
