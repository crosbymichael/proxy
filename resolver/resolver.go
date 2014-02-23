package resolver

import (
	"errors"
	"net"
)

var (
	ErrResolverExists  = errors.New("resolver already exists for key")
	ErrNoRecordInCache = errors.New("no active result in cache")

	Resolvers map[string]Resolver
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
	if Resolvers == nil {
		Resolvers = make(map[string]Resolver)
	}
	if _, exists := Resolvers[name]; exists {
		return ErrResolverExists
	}
	Resolvers[name] = r

	return nil
}
