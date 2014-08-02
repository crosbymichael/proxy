package proxy

import (
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/rcrowley/go-metrics"
)

type worker struct {
	sync.Mutex
	p     *tcpProxy
	group *sync.WaitGroup

	config *tls.Config
	closed bool

	liveConns  metrics.Counter
	totalConns metrics.Counter
}

func newWorker(p *tcpProxy, config *tls.Config) *worker {
	w := &worker{
		p:      p,
		config: config,
		group:  &sync.WaitGroup{},
	}

	w.registerCounters()

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
