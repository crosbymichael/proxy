package main

import (
	"flag"
	"github.com/crosbymichael/log"
	"github.com/crosbymichael/proxy"
	"github.com/crosbymichael/proxy/resolver"
	"github.com/crosbymichael/proxy/stats"
	"os"
	"sync"
	"time"
)

var (
	config string
)

func init() {
	flag.StringVar(&config, "c", "config.toml", "config file path")
	flag.Parse()
}

func fatal(format string, err error) {
	log.Logf(log.FATAL, format, err)
	os.Exit(1)
}

func registerResolvers(host *proxy.Host) error {
	if host.Dns != "" {
		if err := resolver.AddResolver("srv", resolver.NewSrvResolver(host.Dns)); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	f, err := os.Open(config)
	if err != nil {
		fatal("open config file %s", err)
	}
	config, err := proxy.LoadConfig(f)
	if err != nil {
		fatal("reading config file %s", err)
	}
	if err := registerResolvers(config); err != nil {
		fatal("registering resolvers %s", err)
	}

	log.Logf(log.DEBUG, "starting metrics")
	stats.Collect(30*time.Second, true)

	log.Logf(log.INFO, "configuration loaded")
	group := &sync.WaitGroup{}
	// TODO: send close to other backends
	for name, backend := range config.Backends {
		group.Add(1)

		var (
			nv = name
			bv = backend
		)
		log.Logf(log.INFO, "starting proxy %s for %s", bv.Proto, nv)
		p, err := proxy.NewProxy(config, bv)
		if err != nil {
			fatal("failed to create proxy %s", err)
		}
		rsv := resolver.Resolvers[bv.Resolver]

		handler, err := proxy.NewHandler(config, bv, rsv)
		if err != nil {
			fatal("failed to create handler %s", err)
		}
		go func() {
			defer group.Done()
			if err := p.Run(handler); err != nil {
				fatal("running proxy %s", err)
			}
		}()
	}
	group.Wait()
	log.Logf(log.INFO, "proxy going down")
}
