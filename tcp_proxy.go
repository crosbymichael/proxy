package proxy

import (
	"github.com/crosbymichael/log"
	"github.com/crosbymichael/proxy/stats"
	"net"
	"time"
)

func newTcpPRoxy(host *Host, backend *Backend) (*tcpProxy, error) {
	return &tcpProxy{
		host:    host,
		backend: backend,
	}, nil
}

type tcpProxy struct {
	backend  *Backend
	host     *Host
	listener *net.TCPListener
}

func (p *tcpProxy) Close() error {
	return p.listener.Close()
}

func (p *tcpProxy) Run(handler Handler) (err error) {
	p.listener, err = net.ListenTCP("tcp", &net.TCPAddr{IP: p.backend.IP, Port: p.backend.Port})
	if err != nil {
		return err
	}
	defer p.Close()

	var (
		errorCount  int
		connections = make(chan *net.TCPConn, p.backend.ConnectionBuffer)
	)

	for i := 0; i < p.backend.MaxConcurrent; i++ {
		go proxyWorker(connections, p.backend, handler)
	}

	for {
		conn, err := p.listener.AcceptTCP()
		if err != nil {
			errorCount++
			if errorCount > p.host.MaxListenErrors {
				return err
			}
			log.Logf(log.ERROR, "tcp accept error %s", err)
			continue
		}
		connections <- conn
	}
	return nil
}

func proxyWorker(c chan *net.TCPConn, backend *Backend, handler Handler) {
	for conn := range c {
		stats.ActiveConnections.Inc(1)
		start := time.Now()
		if err := handler.HandleConn(conn); err != nil {
			log.Logf(log.ERROR, "handle connection %s", err)
		}
		stats.ReqeustTimer.Update(time.Now().Sub(start))
		stats.ActiveConnections.Dec(1)
	}
}
