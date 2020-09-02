package tcp

const (
	FIN ControlFlag = 0x01 // 000001
	SYN ControlFlag = 0x02 // 000010
	RST ControlFlag = 0x04 // 000100
	PSH ControlFlag = 0x08 // 001000
	ACK ControlFlag = 0x10 // 010000
	URG ControlFlag = 0x20 // 100000
	ECN ControlFlag = 0x40 // 1000000
	CWR ControlFlag = 0x80 // 10000000
)

const (
	End OptionKind = 0
	Nop OptionKind = 1
	MSS OptionKind = 2
	WS  OptionKind = 3
	SP  OptionKind = 4
	SCK OptionKind = 5
	TS  OptionKind = 8
)
