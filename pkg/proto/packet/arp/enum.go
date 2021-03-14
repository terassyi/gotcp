package arp

const ARPHeaderSize int = 8

const HARDWARE_ETHERNET HardwareType = 1

const (
	PROTOCOL_IPv4 ProtocolType = 0x0800
	PROTOCOL_IPv6 ProtocolType = 0x86dd
)

const (
	ARP_REQUEST OperationCode = 0x0001
	ARP_REPLY   OperationCode = 0x0002
)
