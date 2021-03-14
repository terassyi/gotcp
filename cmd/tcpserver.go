package cmd

import (
	"context"
	"flag"

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

	var errChan chan error

	go func() {
		buf := make([]byte, 20480)
		l, err := conn.Read(buf)
		if err != nil {
			errChan <- err
			return
		}
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Infof("message recv %d bytes\n", l)
		logrus.Println("message recv> ", string(buf))
		message := "Hello from gotcp server"
		l, err = conn.Write([]byte(message))
		if err != nil {
			errChan <- err
			return
		}
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Infof("message send %dbytes\n", l)
		logrus.Println("message send> ", message)
	}()

	err = <-errChan
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": "tcp server",
		}).Error(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
