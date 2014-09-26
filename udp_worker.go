package proxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

type udpWorker struct {
	sync.Mutex
	p     *udpProxy
	group *sync.WaitGroup

	closed bool

	liveConns  metrics.Counter
	totalConns metrics.Counter
}

func newUdpWorker(p *udpProxy) *udpWorker {
	w := &udpWorker{
		p:     p,
		group: &sync.WaitGroup{},
	}

	w.registerCounters()

	return w
}

func (w *udpWorker) registerCounters() {
	w.liveConns = metrics.NewCounter()
	w.totalConns = metrics.NewCounter()

	metrics.Register(fmt.Sprintf("%s_udp_live_connections", w.p.backend.Name), w.liveConns)
	metrics.Register(fmt.Sprintf("%s_udp_total_connections", w.p.backend.Name), w.totalConns)
}

func (w *udpWorker) unregisterCounters() {
	metrics.Unregister(fmt.Sprintf("%s_udp_live_connections", w.p.backend.Name))
	metrics.Unregister(fmt.Sprintf("%s_udp_total_connections", w.p.backend.Name))
}

// work processes connections from the channel until it is closed
func (w *udpWorker) work() {
	if err := w.handleConn(w.p.conn); err != nil {
		logger.WithField("error", err).Errorf("handle connection")
	}

	w.closed = true

	w.unregisterCounters()

	// signal to the proxy that we are done processing all open connections
	w.p.group.Done()
}

func (w *udpWorker) handleConn(rawConn net.Conn) error {
	conn, ok := rawConn.(*net.UDPConn)
	if !ok {
		return fmt.Errorf("invalid net.Conn, not udp")
	}

	start := time.Now()

	udpLiveConnections.Inc(1)
	w.liveConns.Inc(1)
	w.totalConns.Inc(1)

	defer func() {
		conn.Close()

		udpLiveConnections.Dec(1)
		w.liveConns.Dec(1)
	}()

	dest, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   w.p.backend.IP,
		Port: w.p.backend.Port,
	})
	if err != nil {
		return err
	}

	w.group.Add(2)

	c2 := &udpConn{
		readCon:  conn,
		closeCon: conn,
	}

	d2 := &udpConn{
		readCon:  dest,
		closeCon: dest,
	}

	go transferUdp(c2, d2, w.group)
	go transferUdp(d2, c2, w.group)

	w.group.Wait()

	udpRequestTimer.UpdateSince(start)

	return nil
}
