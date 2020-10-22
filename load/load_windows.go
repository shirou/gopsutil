// +build windows

package load

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/internal/common"
)

var (
	loadErr              error
	loadAvg1M            float64 = 0.0
	loadAvg5M            float64 = 0.0
	loadAvg15M           float64 = 0.0
	loadAvgMutex         sync.RWMutex
	loadAvgGoroutineOnce sync.Once
)

// loadAvgGoroutine updates avg data by fetching current load by interval
// TODO register callback rather than this
// see https://psutil.readthedocs.io/en/latest/#psutil.getloadavg
// code https://github.com/giampaolo/psutil/blob/master/psutil/arch/windows/wmi.c
func loadAvgGoroutine() {
	const (
		loadAvgFactor1F  = 0.9200444146293232478931553241
		loadAvgFactor5F  = 0.9834714538216174894737477501
		loadAvgFactor15F = 0.9944598480048967508795473394
	)

	var (
		interval = 5 * time.Second
	)

	generator, err := common.NewProcessorQueueLengthGenerator()
	if err != nil {
		loadAvgMutex.Lock()
		loadErr = err
		loadAvgMutex.Unlock()
		return
	}
	for {
		time.Sleep(interval)
		if generator == nil {
			return
		}
		currentLoad, err := generator.Generate()
		loadAvgMutex.Lock()
		loadErr = err
		if err == nil {
			loadAvg1M = loadAvg1M*loadAvgFactor1F + currentLoad*(1.0-loadAvgFactor1F)
			loadAvg5M = loadAvg5M*loadAvgFactor5F + currentLoad*(1.0-loadAvgFactor5F)
			loadAvg15M = loadAvg15M*loadAvgFactor15F + currentLoad*(1.0-loadAvgFactor15F)
		}
		loadAvgMutex.Unlock()
	}
}

func Avg() (*AvgStat, error) {
	return AvgWithContext(context.Background())
}

func AvgWithContext(ctx context.Context) (*AvgStat, error) {
	loadAvgGoroutineOnce.Do(func() { go loadAvgGoroutine() })
	loadAvgMutex.RLock()
	defer loadAvgMutex.RUnlock()
	ret := AvgStat{
		Load1:  loadAvg1M,
		Load5:  loadAvg5M,
		Load15: loadAvg15M,
	}

	return &ret, loadErr
}

func Misc() (*MiscStat, error) {
	return MiscWithContext(context.Background())
}

func MiscWithContext(ctx context.Context) (*MiscStat, error) {
	ret := MiscStat{}

	return &ret, common.ErrNotImplementedError
}
