package tcp

import "time"

type retransmissionPacket struct {
	timeStamp time.Time
	ackNum uint32
	packet *AddressedPacket
}

type retransmissionFunc func(queue []retransmissionPacket) error