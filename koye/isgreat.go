package main

import (
	"flag"
	"github.com/crosbymichael/proxy"
	"log"
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
		log.Fatal(err)
	}

	config, err := proxy.LoadConfig(f)
	if err != nil {
		log.Fatal(err)
	}

	group := &sync.WaitGroup{}
	for _, backend := range config.Backends {
		group.Add(1)
		go func() {
			defer group.Done()

			if err := proxy.ProxyConnections(backend); err != nil {
				log.Fatal(err)
			}
		}()
	}
	group.Wait()
}
