//go:build aix && !cgo
// +build aix,!cgo

package mem

import (
	"context"

	"github.com/shirou/gopsutil/v3/internal/common"
)

func VirtualMemoryWithContext(ctx context.Context) (*VirtualMemoryStat, error) {
	return nil, common.ErrNotImplementedError
}

func SwapMemoryWithContext(ctx context.Context) (*SwapMemoryStat, error) {
	return nil, common.ErrNotImplementedError
}
