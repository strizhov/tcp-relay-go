package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"github.com/hashicorp/yamux"
)

var (
	net_buf_size int = 4096
)

func main() {
	// Parse
	flag.Parse()
	relayhost := flag.Arg(0)
	relayport := flag.Arg(1)

	// Check arguments
	if relayhost == "" || relayport == "" {
		fmt.Println("usage: ./echoserver <relay host> <relay port>")
		os.Exit(1)
	}

	addrport := relayhost + ":" + relayport
	relayconn, err := net.Dial("tcp", addrport)
	if err != nil {
		fmt.Println("Unable to connect to", addrport)
	}
	defer relayconn.Close()

	// Setup client side of yamux
	session, err := yamux.Client(relayconn, nil)
	if err != nil {
		fmt.Println("Unable to establish yamux client")
		return
	}

	// Open a new stream
	stream, err := session.Open()
	if err != nil {
		fmt.Println("Unable to open yamux stream session")
		return
	}

	buf := make([]byte, 128)
	stream.Read(buf)
	stream.Close()
	
	// Print listening hostname:port
	fmt.Println(string(buf))

	// Accept new connection and echo back messages
	for {
		stream, err := session.Accept()
		if err != nil {
			fmt.Println("Unable to accept new stream")
			break
		}

		go handleNewStream(stream)
	}
}

func handleNewStream(stream net.Conn) {

	defer stream.Close()

	for {
		// Read from stream
		buf := make([]byte, net_buf_size)
		n, err := stream.Read(buf)
		if err != nil{
			fmt.Println("Unable to read the data")
			break
		}

		// Write to stream
		n, err = stream.Write(buf[0:n])
		if err != nil {
			fmt.Println("Unable to write the data")
			break
		}
	}
}

