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
	"github.com/samalba/dockerclient"
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

	var (
		err    error
		client *dockerclient.DockerClient
	)

	if docker != "" {
		client, err = dockerclient.NewDockerClient(docker)
		if err != nil {
			logger.WithField("error", err).Fatal("connecting to docker")
		}
	}

	s := server.New(logger, client)
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

	if err := http.ListenAndServe(":3131", s); err != nil {
		logger.WithField("error", err).Fatal("serving http")
	}
}
