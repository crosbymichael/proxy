package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

type tcpWorker struct {
	sync.Mutex
	p     *tcpProxy
	group *sync.WaitGroup

	config *tls.Config
	closed bool

	liveConns  metrics.Counter
	totalConns metrics.Counter
}

func newTcpWorker(p *tcpProxy, config *tls.Config) *tcpWorker {
	w := &tcpWorker{
		p:      p,
		config: config,
		group:  &sync.WaitGroup{},
	}

	w.registerCounters()

	return w
}

func (w *tcpWorker) registerCounters() {
	w.liveConns = metrics.NewCounter()
	w.totalConns = metrics.NewCounter()

	metrics.Register(fmt.Sprintf("%s_tcp_live_connections", w.p.backend.Name), w.liveConns)
	metrics.Register(fmt.Sprintf("%s_tcp_total_connections", w.p.backend.Name), w.totalConns)
}

func (w *tcpWorker) unregisterCounters() {
	metrics.Unregister(fmt.Sprintf("%s_tcp_live_connections", w.p.backend.Name))
	metrics.Unregister(fmt.Sprintf("%s_tcp_total_connections", w.p.backend.Name))
}

// work processes connections from the channel until it is closed
func (w *tcpWorker) work() {
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

func (w *tcpWorker) handleConn(rawConn net.Conn) error {
	conn, ok := rawConn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("invalid net.Conn, not tcp")
	}

	start := time.Now()

	tcpLiveConnections.Inc(1)
	w.liveConns.Inc(1)
	w.totalConns.Inc(1)

	defer func() {
		conn.CloseRead()
		conn.CloseWrite()
		conn.Close()

		tcpLiveConnections.Dec(1)
		w.liveConns.Dec(1)
	}()

	dest, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   w.p.backend.IP,
		Port: w.p.backend.Port,
	})
	if err != nil {
		return err
	}

	w.group.Add(2)

	c2 := &tcpConn{
		readCon:  conn,
		closeCon: conn,
	}

	if w.config != nil {
		c2.readCon = tls.Server(conn, w.config)
	}

	d2 := &tcpConn{
		readCon:  dest,
		closeCon: dest,
	}

	go transferTcp(c2, d2, w.group)
	go transferTcp(d2, c2, w.group)

	w.group.Wait()

	tcpRequestTimer.UpdateSince(start)

	return nil
}
