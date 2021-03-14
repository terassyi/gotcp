package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/google/subcommands"
	"github.com/sirupsen/logrus"
	"github.com/terassyi/gotcp/pkg/gotcp"
)

type TcpServerCommand struct {
	Iface string
	Port  int
	Debug bool
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
	f.BoolVar(&s.Debug, "debug", false, "output debug message")
}

func (s *TcpServerCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if s.Debug {
		logrus.WithFields(
			logrus.Fields{
				"command": "tcp server",
			}).Debug("debug flag is set")
	} else {
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Info("debug flag is not set")
	}

	// tcp server
	gt, err := gotcp.TcpInit(s.Iface, s.Debug)
	logrus.WithFields(logrus.Fields{
		"command": "tcp client",
	}).Infof("tcp server running at %d\n", s.Port)
	listener, err := gt.Listen("0.0.0.0", s.Port)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Error(err)
		return subcommands.ExitFailure
	}
	conn, err := listener.Accept()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Error(err)
		return subcommands.ExitFailure
	}

	// go func() {
	fmt.Printf("Server> Connection from %v\n", conn.Peer.String())
	buf := ""
	for {
		b := make([]byte, 1448)
		n, err := conn.Read(b)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Server> Connection close by peer\n")
				break
			}
			panic(err)
		}
		fmt.Printf("Server> Read %d bytes\n", n)
		buf += string(b)
		if len(buf) >= 20480 {
			fmt.Printf("Server> recv all buf %d bytes\n", len(buf))
			break
		}

	}
	n, err := conn.Write([]byte(buf))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Server> Write %d bytes\n", n)
	time.Sleep(20 * time.Second)
	fmt.Printf("Server> close\n")
	conn.Close()
	// }()
	return subcommands.ExitSuccess
}
