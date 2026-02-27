// SPDX-License-Identifier: BSD-3-Clause
//go:build !aix

package cpu

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func aixPercent(_ context.Context, _ bool) ([]float64, error) {
	return nil, common.ErrNotImplementedError
}
