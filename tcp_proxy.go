package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
)

type tcpProxy struct {
	backend     *Backend
	listener    net.Listener
	connections chan net.Conn
	group       *sync.WaitGroup
	started     bool
}

func newTcpPRoxy(backend *Backend) (*tcpProxy, error) {
	return &tcpProxy{
		backend:     backend,
		connections: make(chan net.Conn, backend.ConnectionBuffer),
		group:       &sync.WaitGroup{},
	}, nil
}

func (p *tcpProxy) Close() error {
	close(p.connections)

	p.group.Wait()

	err := p.listener.Close()
	p.started = false

	return err
}

func (p *tcpProxy) Backend() *Backend {
	return p.backend
}

func (p *tcpProxy) Start() (err error) {
	if p.started {
		return fmt.Errorf("proxy has already been started")
	}

	p.started = true

	var config *tls.Config
	if p.backend.Cert != "" {
		if config, err = createTLSConfig(p.backend.Cert, p.backend.Key, p.backend.CA); err != nil {
			return err
		}
	}

	if p.listener, err = net.ListenTCP("tcp", &net.TCPAddr{
		IP:   p.backend.BindIP,
		Port: p.backend.BindPort,
	}); err != nil {
		return err
	}

	for i := 0; i < p.backend.MaxConcurrent; i++ {
		logger.Infof("starting worker %d", i)

		p.group.Add(1)

		worker := newWorker(p, config)
		go worker.work()
	}

	go func() {
		for {
			if !p.started {
				break
			}

			conn, err := p.listener.Accept()
			if err != nil {
				logger.WithField("error", err).Errorf("tcp accept")

				continue
			}

			p.connections <- conn
		}
	}()

	return nil
}
