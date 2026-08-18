// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd || openbsd

package process

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func (*Process) EnvironWithContext(_ context.Context) ([]string, error) {
	return nil, common.ErrNotImplementedError
}
