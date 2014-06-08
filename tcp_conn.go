package proxy

import (
	"net"
	"time"
)

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
