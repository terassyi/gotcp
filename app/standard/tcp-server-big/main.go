package main

import (
	"fmt"
	"io"
	"net"
	//"time"
)

func main() {
	listener , err := net.Listen("tcp", "0.0.0.0:8888")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server> running at 0.0.0.0:8888")
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			//defer conn.Close()
			//fmt.Printf("Server> Connection from %v\n", conn.RemoteAddr())
			//buf := make([]byte, 30000)
			//n, err := conn.Read(buf)
			//if err != nil {
			//	panic(err)
			//}
			//fmt.Printf("Clinet> %d\n", n)
			//
			//if _, err := conn.Write(buf[:n]); err != nil {
			//	panic(err)
			//}
			//fmt.Printf("Server> %s\n", string(buf))
			//
			//time.Sleep(time.Second * 30)
			//fmt.Println("Server> connection close.")
			fmt.Printf("Server> Connection from %v\n", conn.RemoteAddr())
			buf := make([]byte, 20000)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					if err == io.EOF {
						fmt.Printf("Server> Connection close by peer\n")
						break
					}
					panic(err)
				}
				fmt.Printf("Server> Read %v bytes\n", n)
				n, err = conn.Write(buf[:n])
				if err != nil {
					panic(err)
				}
				fmt.Printf("Server> Write %v bytes\n", n)
			}
			fmt.Printf("Server> Close\n")
			conn.Close()
		}()
	}

}