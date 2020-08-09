package proto

type ProtocolBuffer struct {
	Buffer chan []byte
}

func NewProtocolBuffer() *ProtocolBuffer {
	return &ProtocolBuffer{Buffer: make(chan []byte, 128)}
}
