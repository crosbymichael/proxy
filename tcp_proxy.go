package proxy

import "net"

func newTcpPRoxy(host *Host, backend *Backend) (*tcpProxy, error) {
	return &tcpProxy{
		host:    host,
		backend: backend,
	}, nil
}

type tcpProxy struct {
	backend  *Backend
	host     *Host
	listener net.Listener
}

func (p *tcpProxy) Close() error {
	return p.listener.Close()
}

func (p *tcpProxy) Run(handler Handler) (err error) {
	p.listener, err = net.ListenTCP("tcp", &net.TCPAddr{IP: p.backend.ListenIP, Port: p.backend.ListenPort})
	if err != nil {
		return err
	}
	defer p.Close()

	var (
		errorCount  int
		connections = make(chan net.Conn, p.backend.ConnectionBuffer)
	)

	for i := 0; i < p.backend.MaxConcurrent; i++ {
		go proxyWorker(connections, p.backend, handler)
	}

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			errorCount++
			if errorCount > p.host.MaxListenErrors {
				return err
			}

			logger.Errorf("tcp accept error %s", err)

			continue
		}

		connections <- conn
	}

	return nil
}

func proxyWorker(c chan net.Conn, backend *Backend, handler Handler) {
	for conn := range c {
		if err := handler.HandleConn(conn); err != nil {
			logger.Errorf("handle connection %s", err)
		}
	}
}
