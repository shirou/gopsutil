// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package load

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLoadAvg(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		stat, err := parseLoadAvg([]string{"1.50", "2.25", "3.00", "1/100", "12345"})
		require.NoError(t, err)
		assert.Equal(t, 1.5, stat.Load1)
		assert.Equal(t, 2.25, stat.Load5)
		assert.Equal(t, 3.0, stat.Load15)
	})

	t.Run("fewer than three fields", func(t *testing.T) {
		t.Parallel()

		stat, err := parseLoadAvg([]string{"1.50", "2.25"})
		require.ErrorContains(t, err, "expected at least 3 fields")
		assert.Nil(t, stat)
	})

	t.Run("invalid field", func(t *testing.T) {
		t.Parallel()

		stat, err := parseLoadAvg([]string{"1.50", "nope", "3.00"})
		require.Error(t, err)
		assert.Nil(t, stat)
	})
}
