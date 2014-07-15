package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/samalba/dockerclient"
)

func newRawTcpHandler(host *Host, backend *Backend) (*tcpHandler, error) {
	p := &tcpHandler{
		host:       host,
		backend:    backend,
		liveConns:  metrics.NewCounter(),
		totalConns: metrics.NewCounter(),
	}

	metrics.Register(fmt.Sprintf("%s_tcp_live_connections", backend.Name), p.liveConns)
	metrics.Register(fmt.Sprintf("%s_tcp_total_connections", backend.Name), p.totalConns)

	if backend.Cert != "" {
		config, err := createTLSConfig(backend.Cert, backend.Key, backend.CA)
		if err != nil {
			return nil, err
		}

		p.config = config
	}

	if host.Docker != "" {
		docker, err := dockerclient.NewDockerClient(host.Docker)
		if err != nil {
			return nil, err
		}

		p.docker = docker

		go p.checkLoop()
	}

	return p, nil
}

type tcpHandler struct {
	sync.Mutex

	host               *Host
	backend            *Backend
	config             *tls.Config
	docker             *dockerclient.DockerClient
	containerIsRunning bool
	closed             bool
	liveConns          metrics.Counter
	totalConns         metrics.Counter
	lastCount          int64
}

func (p *tcpHandler) Close() error {
	metrics.Unregister(fmt.Sprintf("%s_tcp_live_connections", p.backend.Name))
	metrics.Unregister(fmt.Sprintf("%s_tcp_total_connections", p.backend.Name))

	p.closed = true

	return nil
}

func (p *tcpHandler) HandleConn(rawConn net.Conn) error {
	conn, ok := rawConn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("invalid net.Conn, not tcp")
	}

	start := time.Now()

	tcpLiveConnections.Inc(1)
	p.liveConns.Inc(1)
	p.totalConns.Inc(1)

	defer func() {
		conn.CloseRead()
		conn.CloseWrite()
		conn.Close()

		tcpLiveConnections.Dec(1)
		p.liveConns.Dec(1)
	}()

	if !p.containerIsRunning && p.backend.Container != "" {
		if err := p.startContainer(); err != nil {
			return err
		}
	}

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

func (p *tcpHandler) startContainer() error {
	p.Lock()
	defer p.Unlock()

	if !p.containerIsRunning {
		logger.Info("starting container")

		p.containerIsRunning = true

		if err := p.docker.StartContainer(p.backend.Container, nil); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (p *tcpHandler) checkLoop() {
	seconds := p.backend.ContainerStopTimeout
	if seconds <= 0 {
		seconds = 300
	}

	for _ = range time.Tick(time.Duration(seconds) * time.Second) {
		if p.closed {
			return
		}

		current := p.totalConns.Count()

		if p.lastCount == current && p.liveConns.Count() == 0 {
			p.Lock()

			p.containerIsRunning = false
			if err := p.docker.StopContainer(p.backend.Container, 10); err != nil {
				logger.WithField("error", err).Error("stopping container")
			}

			p.Unlock()
		}

		p.lastCount = current
	}
}

func transfer(from, to *tcpConn, group *sync.WaitGroup) {
	if _, err := io.Copy(to, from); err != nil {
		if err, ok := err.(*net.OpError); ok && err.Err == syscall.EPIPE {
			from.CloseWrite()
		}
		logger.Errorf("unexpected error type %s", err)
	}

	to.CloseRead()

	group.Done()
}
