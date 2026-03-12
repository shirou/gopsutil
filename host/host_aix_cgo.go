// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && cgo

package host

import (
	"context"

	"github.com/power-devops/perfstat"
)

func numProcs(_ context.Context) (uint64, error) {
	procs, err := perfstat.ProcessStat()
	if err != nil {
		return 0, err
	}
	return uint64(len(procs)), nil
}
