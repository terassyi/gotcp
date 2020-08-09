package icmp

const (
	EchoReply              Type = 0
	DestinationUnreachable Type = 3
	SourceQuench           Type = 4
	Redirect               Type = 5
	Echo                   Type = 8
	RouterAdvertisement    Type = 9
	RouterSolicitation     Type = 10
	TimeExceeded           Type = 11
	ParameterProblem       Type = 12
	Timestamp              Type = 13
	TimestampReply         Type = 14
	InformationRequest     Type = 15
	InformationReply       Type = 16
	AddressMaskRequest     Type = 17
	AddressMaskReply       Type = 18
)

const (
	EchoReplyCode uint8 = 0

	EchoRequestCode uint8 = 0

	DestinationNetworkUnreachableCode  uint8 = 0
	DestinationHostUnreachableCode     uint8 = 1
	DestinationProtocolUnreachableCode uint8 = 2
	DestinationPortUnreachableCode     uint8 = 3
	FragmentationRequiredCode          uint8 = 4
	SourceRouteFailedCode              uint8 = 5
	DestinationNetworkUnknownCode      uint8 = 6
	DestinationHostUnknown             uint8 = 7

	TTLExpired           uint8 = 0
	FragmentTimeExceeded uint8 = 1
)
