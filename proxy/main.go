package main

import (
	"flag"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/proxy"
)

var (
	docker  string
	irlimit int64

	logger = logrus.New()
)

func init() {
	flag.StringVar(&docker, "docker", "unix:///var/run/docker.sock", "docker api endpoint")
	flag.Int64Var(&irlimit, "rlimit", 0, "rlimit")

	flag.Parse()
}

func setRlimit() error {
	rlimit := uint64(irlimit)

	if rlimit > 0 {
		var limit syscall.Rlimit
		if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
			return err
		}

		logger.WithFields(logrus.Fields{
			"current": limit.Cur,
			"max":     limit.Max,
		}).Info("rlimits")

		if limit.Cur < rlimit {
			limit.Cur = rlimit

			if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	if err := setRlimit(); err != nil {
		logger.Fatalf("setting rlimit %s", err)
	}

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
