package main

import ( 
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	
	"github.com/strizhov/tcp-relay"
	"github.com/hashicorp/yamux"
)

var (
	max_backservers int = 20
)

func main() {
	
	// Parse args
	flag.Parse()

	port, err := strconv.Atoi(flag.Arg(0))
	if flag.NArg() != 1 || err != nil {
		fmt.Println("usage: ./relay <port>")
		os.Exit(1)
	}
	relay_addrport := ":"+strconv.Itoa(port)
	
	// Create socket
	listener, err := net.Listen("tcp", relay_addrport)
	if err != nil {
		fmt.Println("Failed to open local port to listen")
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on", relay_addrport)
	
	// Allocate socket pool
	backend_port_pool := make(chan int, max_backservers)
	for i := 1; i <= max_backservers; i++ {
		backend_port_pool <- i + port
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection '%s'", err)
			continue
		}

		go handleBackendServer(conn, backend_port_pool)
	}
}

func handleBackendServer(backend_sock net.Conn, backend_port_pool chan int) {
	
	defer backend_sock.Close()

	// Setup server side of yamux
	backend_session, err := yamux.Server(backend_sock, nil)
	if err != nil {
		fmt.Println("Unable to start yamux server")
		return
	}

	// Accept a stream
	backend_stream, err := backend_session.Accept()
	if err != nil {
		fmt.Println("Unable to accept yamux session")
		return
	}
	
	// Get free socket from the pool
	clientport := <-backend_port_pool
	clientaddress := "localhost" + ":" + strconv.Itoa(clientport)
	
	// Send newly created address over network
	byteArray := []byte(clientaddress)
	_, err = backend_stream.Write(byteArray)
	if err != nil {
		fmt.Println("Unable to send data to backend server via stream")
		return
	}
	backend_stream.Close()

	// Create client server
	client_listener, err := net.Listen("tcp", clientaddress)
	if err != nil {
		fmt.Println("Failed to listen client port", clientport)
		return
	}
	
	// If we are good, then close connection later and and return port to pool
	defer (func() {
		client_listener.Close()
		backend_port_pool <- clientport
	})()

	for {
		client_sock, err := client_listener.Accept()
		if err != nil {
			fmt.Println("Unable to accept new client")
			return
		}
		defer client_sock.Close()

		backend_stream, err := backend_session.Open()
		if err != nil {
			fmt.Println("Unable to open new stream to backend server, is backend dead?")
			break
		}
		
		var p *proxy.Proxy
		p = proxy.New(client_sock, backend_stream)
		go p.Start()
	}
}

