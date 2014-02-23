package proxy

import (
	"fmt"
	"github.com/crosbymichael/proxy/resolver"
	"io"
	"net"
)

type Proxy interface {
	io.Closer
	Run(Handler) error
}

type Handler interface {
	HandleConn(net.Conn) error
}

func NewProxy(host *Host, backend *Backend) (proxy Proxy, err error) {
	switch backend.Proto {
	case "tcp", "http":
		proxy, err = newTcpPRoxy(host, backend)
	case "udp":
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported protocol %s", backend.Proto)
	}
	return
}

func NewHandler(host *Host, backend *Backend, rsv resolver.Resolver) (handler Handler, err error) {
	switch backend.Proto {
	case "tcp":
		handler, err = newRawTcpHandler(host, backend, rsv)
	case "http":
		handler, err = newHttpHandler(host, backend, rsv)
	default:
		return nil, fmt.Errorf("unsupported protocol %s", backend.Proto)
	}
	return
}
