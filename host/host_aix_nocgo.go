// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package host

import (
	"context"
	"strconv"
	"strings"
)

func numProcs(ctx context.Context) (uint64, error) {
	out, err := invoke.CommandWithContext(ctx, "sh", "-c", "ps aux | wc -l")
	if err != nil {
		return 0, err
	}
	countStr := strings.TrimSpace(string(out))
	count, err := strconv.ParseUint(countStr, 10, 64)
	if err != nil {
		return 0, err
	}
	// ps aux includes header line, so subtract 1 to get actual process count
	return count - 1, nil
}
