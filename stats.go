package proxy

import (
	"log"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
)

var (
	tcpLiveConnections metrics.Counter
	tcpRequestTimer    metrics.Timer
	udpLiveConnections metrics.Counter
	udpRequestTimer    metrics.Timer
)

func init() {
	tcpLiveConnections = metrics.NewCounter()
	metrics.Register("tcp_live_connections", tcpLiveConnections)

	tcpRequestTimer = metrics.NewTimer()
	metrics.Register("tcp_requests_timer", tcpRequestTimer)

	udpLiveConnections = metrics.NewCounter()
	metrics.Register("udp_live_connections", udpLiveConnections)

	udpRequestTimer = metrics.NewTimer()
	metrics.Register("udp_requests_timer", udpRequestTimer)
}

func CollectStats() {
	metrics.Log(metrics.DefaultRegistry, 60*time.Second, log.New(os.Stderr, "[stats] ", log.LstdFlags))
}
