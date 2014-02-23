package resolver

import (
	"net"
	"strconv"
)

func NewIpResolver() Resolver {
	return &ipResolver{}
}

type ipResolver struct {
}

func (i *ipResolver) Resolve(query string) (*Result, error) {
	host, sp, err := net.SplitHostPort(query)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(sp)
	if err != nil {
		return nil, err
	}
	return &Result{
		IP:   net.ParseIP(host),
		Port: port,
	}, nil
}
