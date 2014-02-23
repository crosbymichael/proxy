package resolver

import (
	"errors"
	"net"
)

var (
	ErrResolverExists   = errors.New("resolver already exists for key")
	ErrResolverNotExist = errors.New("resolver for name does not exist")
	ErrNoRecordInCache  = errors.New("no active result in cache")

	resolvers map[string]Resolver
)

// Result represents the service address for a container
type Result struct {
	IP   net.IP
	Port int
}

// Resolver takes a query and returns the IP and Port for the service
// based on the query
type Resolver interface {
	Resolve(query string) (*Result, error)
}

// AddResolver adds a new resolver for a given key
func AddResolver(name string, r Resolver) error {
	if resolvers == nil {
		resolvers = make(map[string]Resolver)
	}
	if _, exists := resolvers[name]; exists {
		return ErrResolverExists
	}
	resolvers[name] = r

	return nil
}

func GetResolver(name string) (Resolver, error) {
	r, exists := resolvers[name]
	if !exists {
		return nil, ErrResolverNotExist
	}
	return r, nil
}
