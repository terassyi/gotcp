package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "172.20.0.3:8888")
	if err != nil {
		panic(err)
	}

	file, err := os.Open("../../../data/random-data")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf := make([]byte, 20480)
	if _, err := file.Read(buf); err != nil {
		panic(err)
	}

	if _, err := conn.Write(buf); err != nil {
		panic(err)
	}
	fmt.Println("Client> write 2000 bytes to the server")

	res := make([]byte, 20480)
	if _, err := conn.Read(res); err != nil {
		panic(err)
	}
	fmt.Printf("Server> %s \n", string(res))

	if err := conn.Close(); err != nil {
		panic(err)
	}
	fmt.Println("connection close. exit.")
}
