package proxy

import (
	"fmt"
	"github.com/crosbymichael/proxy/resolver"
	"net"
	"net/http"
	"net/http/httputil"
)

func newHttpHandler(host *Host, backend *Backend) (*httpHandler, error) {
	if backend.Proto != "http" {
		return nil, fmt.Errorf("invalid proto of http handler %d", backend.Proto)
	}
	domains := make(map[string]string, len(backend.Domains))
	for name, d := range backend.Domains {
		domains[name] = d.Query
	}
	return &httpHandler{
		domains: domains,
		host:    host,
	}, nil
}

type httpHandler struct {
	host    *Host
	domains map[string]string
}

func (p *httpHandler) HandleConn(rawConn net.Conn) error {
	conn, ok := rawConn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("invalid net.Conn, not tcp")
	}

	serverConn := httputil.NewServerConn(conn, nil)
	defer serverConn.Close()

	request, err := serverConn.Read()
	if err != nil {
		return err
	}

	query, exists := p.domains[request.Host]
	if !exists {
		response := &http.Response{
			StatusCode: http.StatusNotFound,
		}
		serverConn.Write(request, response)
		return fmt.Errorf("host %s is not registered with this proxy", request.Host)
	}

	result, err := resolver.Resolve(query, p.host.Dns)
	if err != nil {
		return err
	}

	dest, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: result.IP, Port: result.Port})
	if err != nil {
		return err
	}
	defer dest.Close()

	clientConn := httputil.NewProxyClientConn(dest, nil)
	defer clientConn.Close()

	response, err := clientConn.Do(request)
	if err != nil {
		return err
	}
	if err := serverConn.Write(request, response); err != nil {
		return err
	}
	return nil
}
