package proxy

import (
	"os"
	"time"

	"log"

	"github.com/rcrowley/go-metrics"
)

var (
	tcpLiveConnections metrics.Counter
	tcpRequestTimer    metrics.Timer
)

func init() {
	tcpLiveConnections = metrics.NewCounter()
	metrics.Register("tcp_live_connections", tcpLiveConnections)

	tcpRequestTimer = metrics.NewTimer()
	metrics.Register("tcp_requests_timer", tcpRequestTimer)
}

func CollectStats() {
	metrics.Log(metrics.DefaultRegistry, 10*time.Second, log.New(os.Stderr, "[stats] ", log.LstdFlags))
}
