package proxy

import (
	"fmt"
	"io"
	"net"
)

type Proxy interface {
	io.Closer
	Run(Handler) error
}

type Handler interface {
	io.Closer
	HandleConn(net.Conn) error
}

func NewProxy(host *Host, backend *Backend) (proxy Proxy, err error) {
	switch backend.Proto {
	case "tcp":
		proxy, err = newTcpPRoxy(host, backend)
	default:
		return nil, fmt.Errorf("unsupported protocol %s", backend.Proto)
	}

	return
}

func NewHandler(host *Host, backend *Backend) (handler Handler, err error) {
	switch backend.Proto {
	case "tcp":
		handler, err = newRawTcpHandler(host, backend)
	default:
		return nil, fmt.Errorf("unsupported protocol %s", backend.Proto)
	}

	return
}
