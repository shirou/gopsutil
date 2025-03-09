// SPDX-License-Identifier: BSD-3-Clause
package load

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestAvg(t *testing.T) {
	v, err := Avg()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)

	empty := &AvgStat{}
	assert.NotSamef(t, v, empty, "error load: %v", v)
	t.Log(v)
}

func TestAvgStat_String(t *testing.T) {
	v := AvgStat{
		Load1:  10.1,
		Load5:  20.1,
		Load15: 30.1,
	}
	e := `{"load1":10.1,"load5":20.1,"load15":30.1}`
	assert.JSONEqf(t, e, fmt.Sprintf("%v", v), "LoadAvgStat string is invalid: %v", v)
	t.Log(e)
}

func TestMisc(t *testing.T) {
	v, err := Misc()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)

	empty := &MiscStat{}
	assert.NotSamef(t, v, empty, "error load: %v", v)
	t.Log(v)
}

func TestMiscStatString(t *testing.T) {
	v := MiscStat{
		ProcsTotal:   4,
		ProcsCreated: 5,
		ProcsRunning: 1,
		ProcsBlocked: 2,
		Ctxt:         3,
	}
	e := `{"procsTotal":4,"procsCreated":5,"procsRunning":1,"procsBlocked":2,"ctxt":3}`
	assert.JSONEqf(t, e, fmt.Sprintf("%v", v), "TestMiscString string is invalid: %v", v)
	t.Log(e)
}

func BenchmarkLoad(b *testing.B) {
	loadAvg := func(tb testing.TB) {
		tb.Helper()
		v, err := Avg()
		common.SkipIfNotImplementedErr(tb, err)
		require.NoErrorf(tb, err, "error %v", err)
		empty := &AvgStat{}
		assert.NotSamef(tb, v, empty, "error load: %v", v)
	}

	b.Run("FirstCall", func(b *testing.B) {
		loadAvg(b)
	})

	b.Run("SubsequentCalls", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			loadAvg(b)
		}
	})
}
