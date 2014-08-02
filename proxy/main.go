package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/proxy"
	"github.com/crosbymichael/proxy/server"
)

var (
	addr    string
	irlimit int64

	logger = logrus.New()
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:3111", "proxy REST API address")
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

	s := server.New(logger)
	go func() {
		sigc := make(chan os.Signal, 10)
		signal.Notify(sigc, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

		for _ = range sigc {
			if err := s.Close(); err != nil {
				logger.WithField("error", err).Fatal("closing server")
			}
			os.Exit(0)
		}
	}()
	go proxy.CollectStats()

	if err := http.ListenAndServe(addr, s); err != nil {
		logger.WithField("error", err).Fatal("serving http")
	}
}
