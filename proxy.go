package proxy

import (
	"fmt"
	"io"
)

type Proxy interface {
	io.Closer
	Run() error
}

func NewProxy(host *Host, backend *Backend) (proxy Proxy, err error) {
	switch backend.Proto {
	case "tcp":
		proxy, err = newTcpPRoxy(host, backend)
	case "udp":
	case "http":
	default:
		return nil, fmt.Errorf("unsupported protocol %s", backend.Proto)
	}
	return
}
