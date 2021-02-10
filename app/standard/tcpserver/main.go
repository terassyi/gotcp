package main

import (
	"fmt"
	"net"
	"io"
)

func main() {

	listener, error := net.Listen("tcp", "0.0.0.0:8888");

	if error != nil {
		panic(error);
	}

	fmt.Println("Server running at 0.0.0.0:8888");

	waitClient(listener);

}

func waitClient(listener net.Listener) {
	connection, error := listener.Accept();

	if error != nil {
		panic(error);
	}

	go goEcho(connection);

	waitClient(listener);
}

func goEcho(connection net.Conn) {
	defer connection.Close();
	echo(connection);
}

func echo(connection net.Conn) {

	var buf = make([]byte, 1024);

	n, error := connection.Read(buf);
	if (error != nil) {
		if error == io.EOF {
			return;
		} else {
			panic(error);
		}
	}

	fmt.Printf("Client> %s \n", buf);

	n, error = connection.Write(buf[:n])
	if error != nil {
		panic(error);
	}

	echo(connection)
}
