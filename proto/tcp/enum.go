package tcp

const (
	CLOSED      state = 0
	LISTEN      state = 1
	SYN_SENT    state = 2
	SYN_RECVD   state = 3
	ESTABLISHED state = 4
	FIN_WAIT1   state = 5
	FIN_WAIT2   state = 6
	CLOSING     state = 7
	TIME_WAIT   state = 8
	CLOSE_WAIT  state = 9
	LAST_ACK    state = 10
)

const windowZero uint16 = 0
