package tcp

type OptionKind uint8

type Option interface {
	Kind() OptionKind
	Length() uint8
	Data() []byte
	Byte() []byte
}

type Options []Option

func NewOptions(data []byte) Options {

}

func (op Options) Byte() []byte {
	var data []byte
	for _, o := range op {
		data = append(data, o.Byte()...)
	}
	return data
}

type EndOfOptionList struct{}

func (EndOfOptionList) Kind() OptionKind {
	return OptionKind(0)
}

func (EndOfOptionList) Length() int {
	return 1
}

func (EndOfOptionList) Data() []byte {
	return nil
}

func (EndOfOptionList) Byte() []byte {
	return []byte{byte(0)}
}

type NoOperation struct{}

func (NoOperation) Kind() OptionKind {
	return OptionKind(1)
}

func (NoOperation) Length() int {
	return 1
}

func (NoOperation) Data() []byte {
	return nil
}

func (NoOperation) Byte() []byte {
	return []byte{byte(1)}
}

type MaxSegmentSize uint16

func (MaxSegmentSize) Kind() OptionKind {
	return OptionKind(2)
}

func (MaxSegmentSize) Length() int {
	return 4
}

func (mss MaxSegmentSize) Data() []byte {
	return []byte{byte(mss >> 8), byte(mss & 0xff)}
}

func (mss MaxSegmentSize) Byte() []byte {
	d := []byte{byte(2), byte(4)}
	return append(d, mss.Data()...)
}

type WindowScale uint8

func (WindowScale) Kind() OptionKind {
	return OptionKind(3)
}

func (WindowScale) Length() int {
	return 3
}

func (ws WindowScale) Data() []byte {
	return []byte{byte(ws)}
}

func (ws WindowScale) Byte() []byte {
	return []byte{byte(3), byte(3), byte(ws)}
}

type SACKPermitted struct{}

func (SACKPermitted) Kind() OptionKind {
	return OptionKind(4)
}

func (SACKPermitted) Length() int {
	return 2
}

func (SACKPermitted) Data() []byte {
	return nil
}

func (sp SACKPermitted) Byte() []byte {
	return []byte{byte(4), byte(2)}
}

type SACK []byte

func (SACK) Kind() OptionKind {
	return OptionKind(5)
}

func (s SACK) Length() int {
	return len(s) + 2
}

func (s SACK) Data() []byte {
	return s
}

func (s SACK) Byte() []byte {
	return append([]byte{byte(5), byte(s.Length())}, s.Data()...)
}

type TimeStamp []byte

func (TimeStamp) Kind() OptionKind {
	return OptionKind(8)
}

func (TimeStamp) Length() int {
	return 10
}

func (t TimeStamp) Data() []byte {
	return t
}

func (t TimeStamp) Byte() []byte {
	return append([]byte{byte(8), byte(10)}, t.Data()...)
}
