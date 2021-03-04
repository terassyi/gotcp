package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	time.Sleep(20 * time.Second)
	conn, err := net.Dial("tcp", "172.20.0.2:8888")
	if err != nil {
		panic(err)
	}

	file, err := os.Open("../../../data/random-data")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf := make([]byte, 30000)
	if _, err := file.Read(buf); err != nil {
		panic(err)
	}

	if _, err := conn.Write(buf); err != nil {
		panic(err)
	}
	fmt.Println("Client> write 2000 bytes to the server")

	time.Sleep(1 * time.Second)
	res := make([]byte, 30000)
	if _, err := conn.Read(res); err != nil {
		panic(err)
	}
	fmt.Printf("Server> %s \n", string(res))

	time.Sleep(20 * time.Second)
	if err := conn.Close(); err != nil {
		panic(err)
	}
	time.Sleep(2 * time.Second)
	fmt.Println("connection close. exit.")
}
