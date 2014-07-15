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

func newWorker(p *tcpProxy, docker *dockerclient.DockerClient, config *tls.Config) *worker {
	w := &worker{
		p:      p,
		docker: docker,
		config: config,
		group:  &sync.WaitGroup{},
	}

	w.registerCounters()

	if w.p.backend.Container != "" {
		go w.checkLoop()
	}

	return w
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
	w.Lock()
	defer w.Unlock()

	if !w.containerIsRunning {
		logger.Info("starting container")

		w.containerIsRunning = true

		if err := w.docker.StartContainer(w.p.backend.Container, nil); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (w *worker) checkLoop() {
	seconds := w.p.backend.ContainerStopTimeout
	if seconds <= 0 {
		seconds = 300
	}

	for _ = range time.Tick(time.Duration(seconds) * time.Second) {
		if w.closed {
			return
		}

		current := w.totalConns.Count()

		if w.lastCount == current && w.liveConns.Count() == 0 {
			w.Lock()

			w.containerIsRunning = false
			if err := w.docker.StopContainer(w.p.backend.Container, 10); err != nil {
				logger.WithField("error", err).Error("stopping container")
			}

			w.Unlock()
		}

		w.lastCount = current
	}
}
