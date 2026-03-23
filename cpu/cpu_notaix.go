// SPDX-License-Identifier: BSD-3-Clause
//go:build !aix

package cpu

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func aixPercent(_ context.Context, _ time.Duration, _ bool) ([]float64, error) {
	return nil, common.ErrNotImplementedError
}
