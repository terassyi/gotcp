package tcp

const (
	FIN ControlFlag = 0x01
	SYN ControlFlag = 0x02
	RST ControlFlag = 0x04
	PSH ControlFlag = 0x08
	ACK ControlFlag = 0x10
	URG ControlFlag = 0x20
)

const (
	CLOSED      cbState = 0
	LISTEN      cbState = 1
	SYN_SENT    cbState = 2
	SYN_RECVD   cbState = 3
	ESTABLISHED cbState = 4
	FIN_WAIT1   cbState = 5
	FIN_WAIT2   cbState = 6
	CLOSING     cbState = 7
	TIME_WAIT   cbState = 8
	CLOSE_WAIT  cbState = 9
	LAST_ACK    cbState = 10
)

const windowZero uint16 = 0
