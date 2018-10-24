package proxy

import (
	"net"
)

var (
	net_buf_size int = 4096
)

// Relay Proxy - Manages a Proxy connection, piping data between local and remote.
type Proxy struct {
	totalBytes	uint64
	lconn, rconn	net.Conn
}

// New - Create a new Relay Proxy instance. Takes over local connection passed in,
// and closes it when finished.
func New(lconn, rconn net.Conn) *Proxy {
	return &Proxy{
		lconn:  lconn,
		rconn:  rconn,
	}
}

// Start - open connection to remote and start proxying data.
func (p *Proxy) Start() {
	// Open two directional pipes
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)
}

func (p *Proxy) pipe(src, dst net.Conn) {
        
	buff := make([]byte, net_buf_size)
	for {
		// Read from source
		n, err := src.Read(buff)
		if err != nil {
			return
		}
		b := buff[:n]

		// Write out result
		n, err = dst.Write(b)
		if err != nil {
			return
		}

		// Collect bytes for stats
		p.totalBytes += uint64(n)
	}
}
