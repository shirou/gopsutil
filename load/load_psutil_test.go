// SPDX-License-Identifier: BSD-3-Clause
package load

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
	pt "github.com/shirou/gopsutil/v4/internal/common/psutiltest"
)

const loadDelta = 0.5 // absolute; loadavg only updates every 5 seconds

func TestAvg_Against_Psutil(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Both sides emulate a load average on windows with background
		// sampling and different warm-up behavior; not comparable.
		t.Skip("load average is emulated on windows")
	}
	pt.RequirePsutil(t)

	_, err := Avg()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		v, err := Avg()
		if !assert.NoError(c, err) {
			return
		}
		py, err := pt.Eval[[]float64](t, "psutil.getloadavg()")
		if !assert.NoError(c, err) {
			return
		}
		if !assert.Len(c, py, 3) {
			return
		}
		assert.NoError(c, pt.CheckWithinTolerance("Load1", py[0], v.Load1, 0, loadDelta))
		assert.NoError(c, pt.CheckWithinTolerance("Load5", py[1], v.Load5, 0, loadDelta))
		assert.NoError(c, pt.CheckWithinTolerance("Load15", py[2], v.Load15, 0, loadDelta))
	}, pt.DefaultTimeout, pt.DefaultTick)
}

func TestMiscCtxt_Against_Psutil(t *testing.T) {
	pt.RequirePsutil(t)

	// ProcsRunning/ProcsBlocked/ProcsCreated are deliberately not compared:
	// they are instantaneous values with constant churn.

	type pyCPUStats struct {
		CtxSwitches uint64 `json:"ctx_switches"`
	}

	before, err := pt.Eval[pyCPUStats](t, "psutil.cpu_stats()")
	require.NoError(t, err)

	v, err := Misc()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	if v.Ctxt == 0 {
		t.Skip("Ctxt not populated on this platform")
	}
	// MiscStat.Ctxt is int; on 32-bit platforms the kernel's cumulative
	// counter can exceed it and wrap. Surface that clearly instead of
	// comparing a garbage conversion.
	require.Positivef(t, v.Ctxt, "Ctxt overflowed int: %d", v.Ctxt)

	after, err := pt.Eval[pyCPUStats](t, "psutil.cpu_stats()")
	require.NoError(t, err)

	pt.AssertBracketed(t, "Ctxt", before.CtxSwitches, uint64(v.Ctxt), after.CtxSwitches)
}
