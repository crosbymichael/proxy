package resolver

import (
	"errors"
	"fmt"
	"github.com/crosbymichael/log"
	"github.com/miekg/dns"
	"net"
	"sync"
	"time"
)

var (
	client    = &dns.Client{}
	cache     = make(map[string][]*Result)
	cacheLock = sync.RWMutex{}

	ErrNoRecordInCache = errors.New("no active result in cache")
)

type Result struct {
	IP        net.IP
	Port      int
	TTL       int
	Active    bool
	TimeAdded time.Time
}

func init() {
	// start the garbage collector
	go gcInactiveItems()
}

func Resolve(query, server string) (*Result, error) {
	result, err := checkCache(query)
	if err != nil {
		if err != ErrNoRecordInCache {
			return nil, err
		}

		reply, err := resolveSRV(query, server)
		if err != nil {
			return nil, err
		}
		if len(reply.Answer) == 0 {
			return nil, fmt.Errorf("no backends avaliable for %s", query)
		}
		result = fillCache(query, reply)
	}
	return result, nil
}

func resolveSRV(query, server string) (*dns.Msg, error) {
	msg := &dns.Msg{}
	msg.SetQuestion(query, dns.TypeSRV)

	reply, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, err
	}
	return reply, err
}

func fillCache(query string, msg *dns.Msg) *Result {
	results := cache[query]
	for i, answer := range msg.Answer {
		if srv, ok := answer.(*dns.SRV); ok {
			extra := msg.Extra[i]
			if ev, ok := extra.(*dns.A); ok {
				r := &Result{
					Port:      int(srv.Port),
					TimeAdded: time.Now(),
					TTL:       int(srv.Header().Ttl),
					Active:    true,
					IP:        ev.A,
				}
				results = append(results, r)
			}
		}
	}
	cache[query] = results
	return results[0]
}

func createResult(msg *dns.Msg) (*Result, error) {
	var (
		first  = msg.Answer[0]
		result = &Result{}
	)
	v, ok := first.(*dns.SRV)
	if !ok {
		return nil, fmt.Errorf("dns answer not valid SRV record")
	}
	result.Port = int(v.Port)
	result.TimeAdded = time.Now()
	result.TTL = int(v.Header().Ttl)
	result.Active = true

	for _, extra := range msg.Extra {
		if ev, ok := extra.(*dns.A); ok {
			result.IP = ev.A
		}
	}
	return result, nil
}

func gcInactiveItems() {
	itemsRemoved := 0
	for _ = range time.Tick(3 * time.Minute) {
		log.Logf(log.DEBUG, "gc started")
		cacheLock.Lock()

		for key, results := range cache {
			cleaned := []*Result{}
			for _, r := range results {
				if r.Active {
					cleaned = append(cleaned, r)
				} else {
					itemsRemoved++
				}
			}
			cache[key] = cleaned
		}
		cacheLock.Unlock()
		log.Logf(log.DEBUG, "gc ended removing %d items", itemsRemoved)
	}
}

func checkCache(query string) (*Result, error) {
	cacheLock.RLock()
	defer cacheLock.RUnlock()

	results, exists := cache[query]
	if !exists {
		return nil, ErrNoRecordInCache
	}

	now := time.Now()
	for _, r := range results {
		if r.Active {
			// expire the result if needed
			if int(now.Sub(r.TimeAdded).Seconds()) > r.TTL {
				r.Active = false
				continue
			}
		}
		return r, nil
	}
	return nil, ErrNoRecordInCache
}
