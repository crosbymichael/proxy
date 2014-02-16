package main

import (
	"flag"
	"github.com/crosbymichael/log"
	"github.com/crosbymichael/proxy"
	"os"
	"sync"
)

var (
	config string
)

func init() {
	flag.StringVar(&config, "c", "config.toml", "config file path")
	flag.Parse()
}

func main() {
	f, err := os.Open(config)
	if err != nil {
		log.Logf(log.FATAL, "%s", err)
		os.Exit(1)
	}

	config, err := proxy.LoadConfig(f)
	if err != nil {
		log.Logf(log.FATAL, "%s", err)
		os.Exit(1)
	}

	group := &sync.WaitGroup{}
	// TODO: send close to other backends
	for name, backend := range config.Backends {
		group.Add(1)

		nv := name
		bv := backend
		go func() {
			defer group.Done()

			log.Logf(log.INFO, "starting proxy for %s", nv)
			if err := proxy.ProxyConnections(bv, config.Dns); err != nil {
				log.Logf(log.FATAL, "%s", err)
				os.Exit(1)
			}
		}()
	}
	group.Wait()
}
