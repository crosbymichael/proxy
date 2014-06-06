package proxy

import (
	"fmt"
	"github.com/crosbymichael/log"
	"github.com/crosbymichael/proxy/resolver"
	"net"
	"net/http"
	"net/http/httputil"
)

func newHttpHandler(host *Host, backend *Backend, rsv resolver.Resolver) (*httpHandler, error) {
	if backend.Proto != "http" {
		return nil, fmt.Errorf("invalid proto of http handler %d", backend.Proto)
	}

	n := len(host.Domains)
	if n == 0 {
		return nil, fmt.Errorf("no domains to register")
	}

	domains := make(map[string]*domain, n)
	for name, d := range host.Domains {
		log.Logf(log.INFO, "adding %s for http proxy", name)
		var (
			nv = d.Domain
			qv = d.Query
			rv = d.Resolver
		)
		rsv, err := resolver.GetResolver(rv)
		if err != nil {
			return nil, err
		}
		domains[nv] = &domain{
			query:    qv,
			resolver: rsv,
		}
	}
	return &httpHandler{
		domains: domains,
		host:    host,
	}, nil
}

type httpHandler struct {
	host    *Host
	domains map[string]*domain
}

type domain struct {
	query    string
	resolver resolver.Resolver
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

	result, err := query.resolver.Resolve(query.query)
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
