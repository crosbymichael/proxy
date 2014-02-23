package resolver

import (
	"fmt"
	"github.com/miekg/dns"
)

func NewSrvResolver(server string) Resolver {
	return &srvResolver{
		server: server,
		client: &dns.Client{},
	}
}

type srvResolver struct {
	server string
	client *dns.Client
}

func (s *srvResolver) Resolve(query string) (*Result, error) {
	reply, err := s.resolveSRV(query)
	if err != nil {
		return nil, err
	}
	if len(reply.Answer) == 0 {
		return nil, fmt.Errorf("no backends avaliable for %s", query)
	}
	result, err := s.createResult(reply)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *srvResolver) resolveSRV(query string) (*dns.Msg, error) {
	msg := &dns.Msg{}
	msg.SetQuestion(query, dns.TypeSRV)

	reply, _, err := s.client.Exchange(msg, s.server)
	if err != nil {
		return nil, err
	}
	return reply, err
}

func (s *srvResolver) createResult(msg *dns.Msg) (*Result, error) {
	var (
		first  = msg.Answer[0]
		result = &Result{}
	)
	v, ok := first.(*dns.SRV)
	if !ok {
		return nil, fmt.Errorf("dns answer not valid SRV record")
	}
	result.Port = int(v.Port)

	for _, extra := range msg.Extra {
		if ev, ok := extra.(*dns.A); ok {
			result.IP = ev.A
		}
	}
	return result, nil
}
