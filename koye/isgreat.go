package main

import (
	"flag"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/proxy"
)

var (
	config string
	logger = logrus.New()
)

func init() {
	flag.StringVar(&config, "c", "config.toml", "config file path")
	flag.Parse()
}

func main() {
	f, err := os.Open(config)
	if err != nil {
		logger.Fatalf("open config file %s", err)
	}

	config, err := proxy.LoadConfig(f)
	if err != nil {
		logger.Fatalf("reading config file %s", err)
	}
	f.Close()

	logger.Infof("configuration loaded")
	group := &sync.WaitGroup{}

	// TODO: send close to other backends
	for name, backend := range config.Backends {
		group.Add(1)

		var (
			nv = name
			bv = backend
		)

		logger.Infof("starting proxy %s for %s", bv.Proto, nv)

		p, err := proxy.NewProxy(config, bv)
		if err != nil {
			logger.Fatalf("failed to create proxy %s", err)
		}

		handler, err := proxy.NewHandler(config, bv)
		if err != nil {
			logger.Fatalf("failed to create handler %s", err)
		}

		go func() {
			defer group.Done()

			if err := p.Run(handler); err != nil {
				logger.Fatalf("running proxy %s", err)
			}
		}()
	}

	group.Wait()

	logger.Infof("proxy going down")
}
