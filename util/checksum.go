package util

func Checksum(data []byte, size int, init uint32) uint16 {
	sum := init
	for i := 0; i < size-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
		if (sum >> 16) > 0 {
			sum = (sum & 0xffff) + (sum >> 16)
		}
	}
	if size&1 == 0 {
		sum += uint32(data[size-1]) << 8
		if (sum >> 16) > 0 {
			sum = (sum & 0xffff) + (sum >> 16)
		}
	}
	return ^(uint16(sum))
}

func Checksum2(data []byte, size int, init uint32) uint16 {
	sum := init
	for i := 0; i < size-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if size&1 != 0 {
		sum += uint32(data[size-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^(uint16(sum))
}
