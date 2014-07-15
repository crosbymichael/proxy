package main

import (
	"flag"
	"os"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/proxy"
)

const MAX_RLIMIT = 10032

var (
	config string
	logger = logrus.New()
)

func init() {
	flag.StringVar(&config, "c", "config.toml", "config file path")
	flag.Parse()
}

func setRlimit() error {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"current": limit.Cur,
		"max":     limit.Max,
	}).Info("rlimits")

	if limit.Cur < MAX_RLIMIT {
		limit.Cur = MAX_RLIMIT

		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
			return err
		}
	}

	return nil
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

	if err := setRlimit(); err != nil {
		logger.Fatalf("setting rlimit %s", err)
	}

	logger.Infof("configuration loaded")
	group := &sync.WaitGroup{}

	// TODO: send close to other backends
	for name, backend := range config.Backends {
		group.Add(1)

		var (
			nv = name
			bv = backend
		)
		bv.Name = nv

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

			handler.Close()
		}()
	}

	go proxy.CollectStats()

	group.Wait()

	logger.Infof("proxy going down")
}
