package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

func (w *worker) handleConn(rawConn net.Conn) error {
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

	if !w.containerIsRunning && w.p.backend.Container != "" {
		if err := w.startContainer(); err != nil {
			return err
		}
	}

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

	go transfer(c2, d2, w.group)
	go transfer(d2, c2, w.group)

	w.group.Wait()

	tcpRequestTimer.UpdateSince(start)

	return nil
}
