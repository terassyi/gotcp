package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "172.20.0.2:8888")
	if err != nil {
		panic(err)
	}

	file, err := os.Open("../../../data/random-data")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	go func() {
		buf := ""
		for {
			b := make([]byte, 1448)
			n, err := conn.Read(b)
			if err != nil {
				if err == io.EOF {
					fmt.Println("Client> Connection closed by peer")
					break
				}
				panic(err)
			}
			fmt.Printf("Client> Read %d bytes\n", n)
			buf += string(b)
			if len(buf) >= 20480 {
				fmt.Printf("Client> recv all %d bytes\n", len(buf))
				break
			}
		}
	}()

	buf := make([]byte, 20480)
	if _, err := file.Read(buf); err != nil {
		panic(err)
	}

	if _, err := conn.Write(buf); err != nil {
		panic(err)
	}
	fmt.Printf("Client> Write %d bytes\n", len(buf))
	time.Sleep(5 * time.Second)

	if err := conn.Close(); err != nil {
		panic(err)
	}
	fmt.Println("connection close. exit.")
}
