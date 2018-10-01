package proxy

import (
	"fmt"
	"io"
	"net"
)

type Proxy interface {
	io.Closer
	Start() error
	Backend() *Backend
}

type handler interface {
	io.Closer
	HandleConn(net.Conn) error
}

func New(backend *Backend) (proxy Proxy, err error) {
	switch backend.Proto {
	case "tcp":
		proxy, err = newTcpProxy(backend)
	case "udp":
		proxy, err = newUdpProxy(backend)
	default:
		return nil, fmt.Errorf("unsupported protocol %s", backend.Proto)
	}

	return
}
