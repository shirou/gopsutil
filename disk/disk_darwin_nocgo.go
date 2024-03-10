// SPDX-License-Identifier: BSD-3-Clause
//go:build (darwin && !cgo) || ios

package disk

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func IOCountersWithContext(ctx context.Context, names ...string) (map[string]IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}
