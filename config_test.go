package proxy

import (
	"strings"
	"testing"
)

var config = `
log = "stderr"
dns = "172.17.42.1:53"

[backends]

    [backends.redis]
    proto = "tcp"
    ip = "192.168.56.9"
    port = 6379
    query = "redis.dev.docker."
    max_concurrent = 20

    [backends.benchmark]
    proto = "tcp"
    ip = "192.168.56.9"
    query = "benchmark.dev.docker."
    port = 8081
    max_concurrent = 100

    [backends.production]
    proto = "http"
    ip = "0.0.0.0"
    port = 80
    max_concurrent = 100

[domains]

    [domains.localhost]
    query = "blog.dev.docker."
`

func TestParseConfig(t *testing.T) {
	r := strings.NewReader(config)

	host, err := LoadConfig(r)
	if err != nil {
		t.Fatal(err)
	}
	if host == nil {
		t.Fatal("host should not be nil")
	}
}
