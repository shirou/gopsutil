package disk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogicalDrives(t *testing.T) {
	ctx := context.Background()
	drives, err := getLogicaldrives(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(drives), 0)
	for _, d := range drives {
		assert.NotEmpty(t, d)
	}
	assert.Contains(t, drives, `C:\`)
}
