package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	listener , err := net.Listen("tcp", "0.0.0.0:8888")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server> running at 0.0.0.0:8888")

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	buf := make([]byte, 1024)
	if _, err := conn.Read(buf); err != nil {
		panic(err)
	}

	fmt.Printf("Clinet> %s\n", string(buf))
	time.Sleep(1 * time.Second)

	file, err := os.Open("../../../data/random-data")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	res := make([]byte, 20000)
	if _, err := file.Read(res); err != nil {
		panic(err)
	}
	if _, err := conn.Write(res); err != nil {
		panic(err)
	}
	fmt.Printf("Server> %s\n", string(res))

	time.Sleep(10 * time.Second)
}