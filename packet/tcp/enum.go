package tcp

const (
	FIN ControlFlag = 0x01 // 000001
	SYN ControlFlag = 0x02 // 000010
	RST ControlFlag = 0x04 // 000100
	PSH ControlFlag = 0x08 // 001000
	ACK ControlFlag = 0x10 // 010000
	URG ControlFlag = 0x20 // 100000
)
