package ping

import (
	"fmt"
	"github.com/terassyi/gotcp/interfaces"
	"github.com/terassyi/gotcp/logger"
	"github.com/terassyi/gotcp/packet/ethernet"
	"github.com/terassyi/gotcp/packet/icmp"
	ippacket "github.com/terassyi/gotcp/packet/ipv4"
	"github.com/terassyi/gotcp/proto"
	"github.com/terassyi/gotcp/proto/arp"
	eth "github.com/terassyi/gotcp/proto/ethernet"
	"github.com/terassyi/gotcp/proto/ipv4"
	"os"
	"time"
)

type Ping struct {
	*proto.ProtocolBuffer
	ip *ipv4.Ipv4
	//icmp *Icmp
	dst      ippacket.IPAddress
	ident    int
	seqNo    int
	SendTime []time.Time
	queue    chan struct{}
	logger   *logger.Logger
}

func New(name, dst string, debug bool) (*Ping, error) {
	addr, err := ippacket.StringToIPAddress(dst)
	if err != nil {
		return nil, err
	}
	iface, err := interfaces.New(name, "afpacket")
	if err != nil {
		return nil, err
	}
	a := arp.New(arp.NewTable(), debug)
	if err := a.SetAddr(name); err != nil {
		return nil, err
	}
	e, err := eth.New(iface, a)
	if err != nil {
		return nil, err
	}
	ip, err := ipv4.New(e, nil, nil, debug)
	if err != nil {
		return nil, err
	}
	return &Ping{
		ProtocolBuffer: proto.NewProtocolBuffer(),
		ip:             ip,
		dst:            *addr,
		ident:          os.Getpid(),
		seqNo:          0,
		SendTime:       make([]time.Time, 0, 128),
		queue:          make(chan struct{}),
		logger:         logger.New(debug, "icmp"),
	}, nil
}

func (p *Ping) Start() error {
	go p.pingSend()
	go p.pingRecv()
	go p.ip.Eth.Arp.Handle()

	p.logger.Info("gotcp ping start")
	//p.ip.Show()
	p.logger.Info("pid: %d\n", p.ident)
	p.logger.Info("dest: %s\n", p.dst)

	if err := p.start(); err != nil {
		return err
	}

	return nil
}

func (p *Ping) start() error {

	defer p.ip.Eth.Close()

	p.queue <- struct{}{}

	for {
		buf := make([]byte, 512)
		_, err := p.ip.Eth.Recv(buf)
		if err != nil {
			return err
		}
		frame, err := ethernet.New(buf)
		if err != nil {
			return err
		}
		switch frame.Type() {
		case ethernet.ETHER_TYPE_ARP:
			p.ip.Eth.Arp.Recv(frame.Payload())
		case ethernet.ETHER_TYPE_IP:
			p.Buffer <- frame.Payload()
		}
	}
}

func (p *Ping) pingRecv() {
	for {
		buf, ok := <-p.Buffer
		if ok {
			ipPacket, err := ippacket.New(buf)
			if err != nil {
				continue
			}
			packet, err := icmp.New(ipPacket.Data)
			if err != nil {
				continue
			}
			switch packet.Header.Type {
			case icmp.EchoReply:
				message, err := icmp.NewEchoMessage(packet.Data)
				if err != nil {
					p.logger.Errorf("failed to encode echo payload: %v\n", err)
					continue
				}
				if len(p.SendTime) <= int(message.Seq) {
					p.logger.Errorf("invalid seq: %d\n", message.Seq)
					continue
				}
				sendTime := p.SendTime[int(message.Seq)].UnixNano() / int64(time.Microsecond)
				recvTime := time.Now().UnixNano() / int64(time.Microsecond)
				sec := recvTime - sendTime

				fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%f ms\n",
					ipPacket.Header.Length,
					ipPacket.Header.Src.String(),
					message.Seq,
					ipPacket.Header.TTL,
					sec)

				time.Sleep(time.Second)
				p.queue <- struct{}{}
			}
		}
	}
}

func (p *Ping) pingSend() {
	for {
		_, ok := <-p.queue
		if ok {
			p.SendTime = append(p.SendTime, time.Now())
			p.seqNo += 1
			message := icmp.EchoMessage{
				Ident: uint16(p.ident),
				Seq:   uint16(p.seqNo),
				Data:  []byte("ping from gotcp"),
			}
			//fmt.Printf("[info] ping message ident=%v, seq=%v\n", message.Ident, message.Seq)
			data, err := message.Serialize()
			if err != nil {
				p.logger.Errorf("failed to serialize icmp echo message: %v", err)
				continue
			}
			req, err := icmp.Build(icmp.Echo, icmp.EchoRequestCode, data)
			if err != nil {
				p.logger.Errorf("failed to build icmp echo request: %v", err)
				continue
			}
			reqBytes, err := req.Serialize()
			if err != nil {
				p.logger.Errorf("failed to serialize icmp echo request: %v", err)
				continue
			}
			if _, err := p.ip.Send(p.dst, ippacket.IPICMPv4Protocol, reqBytes); err != nil {
				p.logger.Errorf("failed to send: %v", err)
				continue
			}
		}
	}
}
