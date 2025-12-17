// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package load

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiscWithContextAIX(t *testing.T) {
	ctx := context.Background()
	misc, err := MiscWithContext(ctx)
	require.NoError(t, err)
	assert.NotNil(t, misc)

	// Process counts should be non-negative
	assert.GreaterOrEqual(t, misc.ProcsTotal, 0)
	assert.GreaterOrEqual(t, misc.ProcsRunning, 0)
	assert.GreaterOrEqual(t, misc.ProcsBlocked, 0)

	// Total should be >= running + blocked
	assert.GreaterOrEqual(t, misc.ProcsTotal, misc.ProcsRunning+misc.ProcsBlocked)
}

func TestMiscAIX(t *testing.T) {
	// Test the non-context version
	misc, err := Misc()
	require.NoError(t, err)
	assert.NotNil(t, misc)

	// Process counts should be non-negative
	assert.GreaterOrEqual(t, misc.ProcsTotal, 0)
	assert.GreaterOrEqual(t, misc.ProcsRunning, 0)
	assert.GreaterOrEqual(t, misc.ProcsBlocked, 0)
}
