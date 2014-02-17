package stats

import (
	"fmt"
	"github.com/crosbymichael/log"
	"io/ioutil"
	"os"
	"runtime"
)

func byteToMb(b uint64) float64 {
	if b <= 0 {
		return 0
	}
	return float64(b) / 1e6
}

func Collect() {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)
	var (
		frees      = memstats.Frees
		goroutines = runtime.NumGoroutine()
		gcs        = memstats.NumGC
		fds        = getFds()
		allocs     = memstats.Alloc
	)

	log.Logf(log.DEBUG, "go routines %d gcs %d fds %d current mem %7.2f MB",
		goroutines, gcs, fds, byteToMb(allocs-frees))
}

func getFds() int {
	if fds, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d/fd", os.Getpid())); err != nil {
		log.Logf(log.ERROR, "Error opening /proc/%d/fd: %s", os.Getpid(), err)
	} else {
		return len(fds)
	}
	return -1
}
