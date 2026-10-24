// SPDX-License-Identifier: BSD-3-Clause
package cpu

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestParseDmesgBoot(t *testing.T) {
	if runtime.GOOS != "freebsd" {
		t.SkipNow()
	}

	cpuTests := []struct {
		file   string
		cpuNum int
		cores  int32
	}{
		{"1cpu_2core.txt", 1, 2},
		{"1cpu_4core.txt", 1, 4},
		{"2cpu_4core.txt", 2, 4},
	}
	for _, tt := range cpuTests {
		v, num, err := parseDmesgBoot(filepath.Join("testdata", "freebsd", tt.file))
		require.NoErrorf(t, err, "parseDmesgBoot failed(%s), %v", tt.file, err)
		assert.Equalf(t, num, tt.cpuNum, "parseDmesgBoot wrong length(%s), %v", tt.file, err)
		assert.Equalf(t, v.Cores, tt.cores, "parseDmesgBoot wrong core(%s), %v", tt.file, err)
		assert.Truef(t, common.StringsContains(v.Flags, "fpu"), "parseDmesgBoot fail to parse features(%s), %v", tt.file, err)
	}
}
