// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootTimeWithContext(t *testing.T) {
	// This is a wrapper function that delegates to common.BootTimeWithContext
	// Actual implementation testing is done in common_aix_test.go
	bootTime, err := BootTimeWithContext(context.TODO())
	require.NoError(t, err)
	assert.Positive(t, bootTime)
}

func TestUptimeWithContext(t *testing.T) {
	// This is a wrapper function that delegates to common.UptimeWithContext
	// Actual implementation testing is done in common_aix_test.go
	uptime, err := UptimeWithContext(context.TODO())
	require.NoError(t, err)
	assert.Positive(t, uptime)
}

func TestUsersWithContext(t *testing.T) {
	// Integration test — reads /etc/utmp directly on a real AIX system.
	// Verifies the binary utmp parsing returns valid user data.
	users, err := UsersWithContext(context.TODO())
	require.NoError(t, err)

	// At least one user should be logged in (us, running this test)
	require.NotEmpty(t, users, "expected at least one logged-in user")

	for _, u := range users {
		assert.NotEmpty(t, u.User, "user name should not be empty")
		assert.NotEmpty(t, u.Terminal, "terminal should not be empty")
		assert.Positive(t, u.Started, "started time should be a positive epoch timestamp")
	}
}
