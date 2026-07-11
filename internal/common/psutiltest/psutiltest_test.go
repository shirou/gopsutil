// SPDX-License-Identifier: BSD-3-Clause
package psutiltest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEval(t *testing.T) {
	RequirePsutil(t)

	// Round-trip of plain python values through the JSON wrapper.
	type pyValues struct {
		A []any   `json:"a"`
		B *string `json:"b"`
	}
	out, err := Eval[pyValues](t, `{"a": [1, 2.5, "x"], "b": None}`)
	require.NoError(t, err)
	require.Len(t, out.A, 3)
	assert.InDelta(t, 1.0, out.A[0], 0.0001)
	assert.InDelta(t, 2.5, out.A[1], 0.0001)
	assert.Equal(t, "x", out.A[2])
	assert.Nil(t, out.B)

	// A psutil scalar.
	boot, err := Eval[float64](t, "psutil.boot_time()")
	require.NoError(t, err)
	assert.Positive(t, boot)

	// A psutil namedtuple decodes via its _asdict() keys.
	type pyVirtualMemory struct {
		Total uint64 `json:"total"`
	}
	vm, err := Eval[pyVirtualMemory](t, "psutil.virtual_memory()")
	require.NoError(t, err)
	assert.Positive(t, vm.Total)
}

func TestEvalFailure(t *testing.T) {
	RequirePsutil(t)

	_, err := Eval[any](t, "1/0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ZeroDivisionError")
}
