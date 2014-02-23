package proxy

import (
	"fmt"
	"github.com/crosbymichael/log"
	"github.com/crosbymichael/proxy/resolver"
	"io"
	"net"
	"sync"
	"syscall"
)

func newRawTcpHandler(host *Host, backend *Backend, rsv resolver.Resolver) (*tcpHandler, error) {
	return &tcpHandler{
		host:     host,
		backend:  backend,
		resolver: rsv,
	}, nil
}

type tcpHandler struct {
	host     *Host
	backend  *Backend
	resolver resolver.Resolver
}

func (p *tcpHandler) HandleConn(rawConn net.Conn) error {
	conn, ok := rawConn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("invalid net.Conn, not tcp")
	}

	defer func() {
		conn.CloseRead()
		conn.CloseWrite()
		conn.Close()
	}()

	result, err := p.resolver.Resolve(p.backend.Query)
	if err != nil {
		return err
	}

	dest, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: result.IP, Port: result.Port})
	if err != nil {
		return err
	}

	group := &sync.WaitGroup{}
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
