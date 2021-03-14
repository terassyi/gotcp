package main

import (
	"encoding/hex"
	"fmt"

	"github.com/terassyi/gotcp/pkg/interfaces"
)

func main() {

	//0000   ce 5c 7f 6b 61 ba 22 ca f6 f3 c3 6e 08 00 45 10   .\.ka."....n..E.
	//0010   00 3c fd 46 40 00 40 06 bc 0f c0 a8 00 02 c0 a8   .<.F@.@.........
	//0020   00 03 eb 1e 00 17 37 08 b7 f1 00 00 00 00 a0 02   ......7.........
	//0030   72 10 81 84 00 00 02 04 05 b4 04 02 08 0a 19 64   r..............d
	//0040   34 32 00 00 00 00 01 03 03 07                     42........

	data := []byte{
		0xce, 0x5c, 0x7f, 0x6b, 0x61, 0xba, 0x22, 0xca, 0xf6, 0xf3, 0xc3, 0x6e, 0x08, 0x00, 0x45, 0x10,
		0x00, 0x3c, 0xfd, 0x46, 0x40, 0x00, 0x40, 0x06, 0xbc, 0x0f, 0xc0, 0xa8, 0x00, 0x02, 0xc0, 0xa8,
		0x00, 0x03, 0xeb, 0x1e, 0x00, 0x17, 0x37, 0x08, 0xb7, 0xf1, 0x00, 0x00, 0x00, 0x00, 0xa0, 0x02,
		0x72, 0x10, 0x81, 0x84, 0x00, 0x00, 0x02, 0x04, 0x05, 0xb4, 0x04, 0x02, 0x08, 0x0a, 0x19, 0x64,
		0x34, 0x32, 0x00, 0x00, 0x00, 0x00, 0x01, 0x03, 0x03, 0x07,
	}
	iface, err := interfaces.New("host1_veth0", "afpacket")
	if err != nil {
		panic(err)
	}
	defer iface.Close()

	fmt.Printf("[info] test tcp syn packet send.")
	l, err := iface.Send(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[info] test tcp syn packet %d byte send.", l)
	rep := make([]byte, 256)
	l, err = iface.Recv(rep)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.Dump(rep))
}
