// SPDX-License-Identifier: BSD-3-Clause
//go:build race

package process

import (
	"errors"
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

			if errors.Is(err, common.ErrNotImplementedError) {
				t.Skip("not implemented")
			}

			require.NoError(t, err, "Ppid() failed, %v", err)

			if j == 9 {
				t.Logf("Ppid(): %d", ppid)
			}
		}(i)
	}
	wg.Wait()
}
