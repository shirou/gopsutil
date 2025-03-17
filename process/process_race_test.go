// SPDX-License-Identifier: BSD-3-Clause
//go:build race

package process

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestPpid_Race(t *testing.T) {
	wg := sync.WaitGroup{}
	testCount := 10
	p := testGetProcess()
	wg.Add(testCount)
	for i := 0; i < testCount; i++ {
		go func(j int) {
			ppid, err := p.Ppid()
			wg.Done()
			common.SkipIfNotImplementedErr(t, err)
			require.NoError(t, err, "Ppid() failed, %v", err)

			if j == 9 {
				t.Logf("Ppid(): %d", ppid)
			}
		}(i)
	}
	wg.Wait()
}
