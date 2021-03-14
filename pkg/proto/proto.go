package proto

type Protocol interface {
	Handle()
}

//type ProtocolHandler struct {
//	ArpTable *arp.Table
//	IpAddress *ipv4.IPAddress
//	MacAddress *ethernet.HardwareAddress
//}
