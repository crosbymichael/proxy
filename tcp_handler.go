package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"
)

func newRawTcpHandler(host *Host, backend *Backend) (*tcpHandler, error) {
	p := &tcpHandler{
		host:    host,
		backend: backend,
	}

	if backend.Cert != "" {
		config, err := createTLSConfig(backend.Cert, backend.Key, backend.CA)
		if err != nil {
			return nil, err
		}

		p.config = config
	}

	return p, nil
}

type tcpHandler struct {
	host    *Host
	backend *Backend
	config  *tls.Config
}

func (p *tcpHandler) HandleConn(rawConn net.Conn) error {
	conn, ok := rawConn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("invalid net.Conn, not tcp")
	}

	start := time.Now()
	tcpLiveConnections.Inc(1)

	defer func() {
		conn.CloseRead()
		conn.CloseWrite()
		conn.Close()
		tcpLiveConnections.Dec(1)
	}()

	dest, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: p.backend.IP, Port: p.backend.Port})
	if err != nil {
		return err
	}

	group := &sync.WaitGroup{}
	group.Add(2)

	c2 := &tcpConn{
		readCon:  conn,
		closeCon: conn,
	}

	if p.config != nil {
		c2.readCon = tls.Server(conn, p.config)
	}

	d2 := &tcpConn{
		readCon:  dest,
		closeCon: dest,
	}

	go transfer(c2, d2, group)
	go transfer(d2, c2, group)

	group.Wait()

	tcpRequestTimer.UpdateSince(start)

	return nil
}

func transfer(from, to *tcpConn, group *sync.WaitGroup) {
	defer group.Done()

	if _, err := io.Copy(to, from); err != nil {
		if err, ok := err.(*net.OpError); ok && err.Err == syscall.EPIPE {
			from.CloseWrite()
		}
		logger.Errorf("unexpected error type %s", err)
	}

	to.CloseRead()
}
