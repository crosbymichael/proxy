package proxy

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/samalba/dockerclient"
)

type worker struct {
	sync.Mutex
	p      *tcpProxy
	docker *dockerclient.DockerClient
	group  *sync.WaitGroup

	config             *tls.Config
	containerIsRunning bool
	closed             bool

	liveConns  metrics.Counter
	totalConns metrics.Counter
	lastCount  int64
}

func newWorker(p *tcpProxy, docker *dockerclient.DockerClient, config *tls.Config) (*worker, error) {
	w := &tcpHandler{
		p:      p,
		docker: docker,
		config: config,
		group:  &sync.WaitGroup{},
	}

	w.registerCounters()

	if w.p.backend.Container != "" {
		go w.checkLoop()
	}

	return p, nil
}

func (w *worker) registerCounters() {
	w.liveConns = metrics.NewCounter()
	w.totalConns = metrics.NewCounter()

	metrics.Register(fmt.Sprintf("%s_tcp_live_connections", w.p.backend.Name), w.liveConns)
	metrics.Register(fmt.Sprintf("%s_tcp_total_connections", w.p.backend.Name), w.totalConns)
}

func (w *worker) unregisterCounters() {
	metrics.Unregister(fmt.Sprintf("%s_tcp_live_connections", w.p.backend.Name))
	metrics.Unregister(fmt.Sprintf("%s_tcp_total_connections", w.p.backend.Name))
}

// work processes connections from the channel until it is closed
func (w *worker) work() {
	for conn := range w.p.connections {
		if err := w.handleConn(conn); err != nil {
			logger.WithField("error", err).Errorf("handle connection")
		}
	}
	w.closed = true

	w.unregisterCounters()

	// signal to the proxy that we are done processing all open connections
	w.p.group.Done()
}

func (w *worker) startContainer() error {
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

func (w *worker) checkLoop() {
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
