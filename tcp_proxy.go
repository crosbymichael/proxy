package proxy

import (
	"github.com/crosbymichael/log"
	"github.com/crosbymichael/proxy/resolver"
	"io"
	"net"
	"sync"
	"syscall"
)

var server = "172.17.42.1:53"

func tcpProxy(c chan *net.TCPConn, backend *Backend) {
	group := &sync.WaitGroup{}
	for conn := range c {
		result, err := resolver.Resolve(backend.Query, server)
		if err != nil {
			log.Logf(log.ERROR, "unable to reslove %s %s", backend.Query, err)
			continue
		}
		if err := handleConnection(conn, group, result); err != nil {
			log.Logf(log.ERROR, "handle connection %s", err)
		}
	}
}

func handleConnection(conn *net.TCPConn, group *sync.WaitGroup, result *resolver.Result) error {
	defer func() {
		conn.CloseRead()
		conn.CloseWrite()
		conn.Close()
	}()

	dest, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: result.IP, Port: result.Port})
	if err != nil {
		return err
	}
	group.Add(2)

	go transfer(conn, dest, group)
	go transfer(dest, conn, group)

	group.Wait()

	return nil
}

func transfer(from, to *net.TCPConn, group *sync.WaitGroup) {
	defer group.Done()
	if _, err := io.Copy(to, from); err != nil {
		if err, ok := err.(*net.OpError); ok && err.Err == syscall.EPIPE {
			from.CloseWrite()
		}
		log.Logf(log.ERROR, "unexpected error type %s", err)
	}
	to.CloseRead()
}

func ProxyConnections(backend *Backend) error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: backend.IP, Port: backend.Port})
	if err != nil {
		return err
	}
	defer listener.Close()

	var (
		errorCount  int
		connections = make(chan *net.TCPConn, backend.ConnectionBuffer)
	)

	for i := 0; i < backend.MaxConcurrent; i++ {
		go tcpProxy(connections, backend)
	}

	// start the main event loop for the proxy
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			errorCount++
			log.Logf(log.ERROR, "tcp accept error %s", err)
			continue
		}
		connections <- conn
	}
	return nil
}
