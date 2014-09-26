package proxy

import (
	"io"
	"net"
	"sync"
	"syscall"
	"time"
)

// transfer transfers bytes from one tcp connection to another
func transferTcp(from, to *tcpConn, group *sync.WaitGroup) {
	if _, err := io.Copy(to, from); err != nil {
		if err, ok := err.(*net.OpError); ok && err.Err == syscall.EPIPE {
			from.CloseWrite()
		}

		logger.Errorf("unexpected error type %s", err)
	}

	to.CloseRead()

	group.Done()
}

// tcpConn is used to handle normal and tls communication
type tcpConn struct {
	readCon  net.Conn
	closeCon *net.TCPConn
}

func (t *tcpConn) Read(b []byte) (int, error) {
	return t.readCon.Read(b)
}

func (t *tcpConn) Write(b []byte) (int, error) {
	return t.readCon.Write(b)
}

func (t *tcpConn) LocalAddr() net.Addr {
	return t.readCon.LocalAddr()
}

func (t *tcpConn) RemoteAddr() net.Addr {
	return t.readCon.RemoteAddr()
}

func (t *tcpConn) SetDeadline(tm time.Time) error {
	return t.readCon.SetDeadline(tm)
}

func (t *tcpConn) SetReadDeadline(tm time.Time) error {
	return t.readCon.SetReadDeadline(tm)
}

func (t *tcpConn) SetWriteDeadline(tm time.Time) error {
	return t.readCon.SetWriteDeadline(tm)
}

func (t *tcpConn) CloseRead() error {
	return t.closeCon.CloseRead()
}

func (t *tcpConn) CloseWrite() error {
	return t.closeCon.CloseWrite()
}

func (t *tcpConn) Close() error {
	return t.closeCon.Close()
}

// transfer transfers bytes from one tcp connection to another
func transferUdp(from, to *udpConn, group *sync.WaitGroup) {
	var buffer [1500]byte
	var err error
	var n int
	if n, err = from.Read(buffer[0:]); err != nil {
		logger.Errorf("unexpected error type %s", err)
	}
	if _, err = to.Write(buffer[0:n]); err != nil {
		logger.Errorf("unexpected error type %s", err)
	}

	group.Done()
}

// tcpConn is used to handle normal and tls communication
type udpConn struct {
	readCon  net.Conn
	closeCon *net.UDPConn
}

func (t *udpConn) Read(b []byte) (int, error) {
	return t.readCon.Read(b)
}

func (t *udpConn) Write(b []byte) (int, error) {
	return t.readCon.Write(b)
}

func (t *udpConn) LocalAddr() net.Addr {
	return t.readCon.LocalAddr()
}

func (t *udpConn) RemoteAddr() net.Addr {
	return t.readCon.RemoteAddr()
}

func (t *udpConn) SetDeadline(tm time.Time) error {
	return t.readCon.SetDeadline(tm)
}

func (t *udpConn) SetReadDeadline(tm time.Time) error {
	return t.readCon.SetReadDeadline(tm)
}

func (t *udpConn) SetWriteDeadline(tm time.Time) error {
	return t.readCon.SetWriteDeadline(tm)
}

func (t *udpConn) Close() error {
	return t.closeCon.Close()
}
