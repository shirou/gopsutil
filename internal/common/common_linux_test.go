//go:build linux
// +build linux

package common

import (
	"context"
	"testing"

	"github.com/shirou/gopsutil/v3/common"
)

func BenchmarkBootTimeWithManyCPUs(b *testing.B) {
	ctx := context.WithValue(context.Background(),
		common.EnvKey,
		common.EnvMap{common.HostProcEnvKey: "testdata/linux/issue_1514"},
	)

	for i := 0; i < b.N; i++ {
		BootTimeWithContext(ctx)
	}
}
