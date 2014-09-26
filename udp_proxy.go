package proxy

import (
	"fmt"
	"net"
	"sync"
)

type udpProxy struct {
	backend *Backend
	conn    *net.UDPConn
	group   *sync.WaitGroup
	started bool
}

func newUdpProxy(backend *Backend) (Proxy, error) {
	return &udpProxy{
		backend: backend,
		group:   &sync.WaitGroup{},
	}, nil
}

func (p *udpProxy) Close() error {
	p.group.Wait()
	err := p.conn.Close()

	return err
}

func (p *udpProxy) Backend() *Backend {
	return p.backend
}

func (p *udpProxy) Start() (err error) {
	if p.started {
		return fmt.Errorf("proxy has already been started")
	}

	p.started = true

	if p.conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   p.backend.BindIP,
		Port: p.backend.BindPort,
	}); err != nil {
		return err
	}

	worker := newUdpWorker(p)
	go worker.work()

	return nil
}
