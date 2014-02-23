package stats

import (
	"github.com/crosbymichael/log"
	"github.com/rcrowley/go-metrics"
	golog "log"
	"os"
	"time"
)

// high level system metrics for the entire proxy
// users can specify more fine grained metrics for http, tcp, etc...
var (
	RequestCount      metrics.Counter // monitor the over number of requests
	RequestErrorCount metrics.Counter // monitor the number of request errors
	ReqeustTimer      metrics.Timer   // monitor request times

	GoRoutines  metrics.Gauge // monitor the number of goroutines
	MemoryGauge metrics.Gauge // monitor the memory usage
	FdGauge     metrics.Gauge // monitor the number of open fds
)

func init() {
	RequestCount = metrics.NewCounter()
	metrics.Register("koye-requests", RequestCount)

	RequestErrorCount = metrics.NewCounter()
	metrics.Register("koye-requests-errors", RequestErrorCount)

	GoRoutines = metrics.NewGauge()
	metrics.Register("koye-goroutines", GoRoutines)

	MemoryGauge = metrics.NewGauge()
	metrics.Register("koye-memory", MemoryGauge)

	FdGauge = metrics.NewGauge()
	metrics.Register("koye-fds", FdGauge)

	ReqeustTimer = metrics.NewTimer()
	metrics.Register("koye-request-time", ReqeustTimer)
}

func Collect(systemTick time.Duration, toStderr bool) error {
	// start collecting system/process information
	go collectSystemInfo(systemTick)

	if toStderr {
		go metrics.Log(metrics.DefaultRegistry, 60e9, golog.New(os.Stderr, "[metrics] ", golog.Lmicroseconds))
	}
	return nil
}

func collectSystemInfo(systemTick time.Duration) {
	var (
		stats *systemInfo
		err   error
	)
	for _ = range time.Tick(systemTick) {
		if stats, err = getSystemInfo(); err != nil {
			log.Logf(log.ERROR, "error getting system info %s", err)
			continue
		}
		GoRoutines.Update(int64(stats.goroutines))
		FdGauge.Update(int64(stats.fds))
		MemoryGauge.Update(stats.memory)
	}
}
