// +build darwin
// +build !cgo

package disk

import "github.com/shirou/gopsutil/internal/common"

func IOCounters() (map[string]IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}
