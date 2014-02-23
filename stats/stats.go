package stats

import (
	"fmt"
	"github.com/crosbymichael/log"
	"io/ioutil"
	"os"
	"runtime"
)

type systemInfo struct {
	goroutines  int
	fds         int
	memory      int64 // in bytes
	numberOfGcs int
}

func getSystemInfo() (*systemInfo, error) {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)

	var (
		frees  = memstats.Frees
		allocs = memstats.Alloc
	)
	return &systemInfo{
		goroutines:  runtime.NumGoroutine(),
		fds:         getFds(),
		memory:      int64(allocs - frees),
		numberOfGcs: int(memstats.NumGC),
	}, nil
}

func getFds() int {
	if fds, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d/fd", os.Getpid())); err != nil {
		log.Logf(log.ERROR, "Error opening /proc/%d/fd: %s", os.Getpid(), err)
	} else {
		return len(fds)
	}
	return -1
}

func byteToMb(b uint64) float64 {
	if b <= 0 {
		return 0
	}
	return float64(b) / 1e6
}
