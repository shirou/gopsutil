// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockInvoker returns canned output for specific commands.
type mockInvoker struct {
	responses map[string]string
}

func (m *mockInvoker) Command(name string, arg ...string) ([]byte, error) {
	key := name + " " + strings.Join(arg, " ")
	key = strings.TrimSpace(key)
	if resp, ok := m.responses[key]; ok {
		return []byte(resp), nil
	}
	return nil, fmt.Errorf("unexpected command: %s", key)
}

func (m *mockInvoker) CommandWithContext(_ context.Context, name string, arg ...string) ([]byte, error) {
	return m.Command(name, arg...)
}

func withMockInvoker(t *testing.T, responses map[string]string) {
	t.Helper()
	old := testInvoker
	testInvoker = &mockInvoker{responses: responses}
	t.Cleanup(func() { testInvoker = old })
}

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

func TestHostIDWithContext(t *testing.T) {
	withMockInvoker(t, map[string]string{
		"uname -u": "IBM,0221D80FV\n",
	})

	id, err := HostIDWithContext(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, "IBM,0221D80FV", id)
}

func TestPlatformInformationWithContext(t *testing.T) {
	withMockInvoker(t, map[string]string{
		"uname -s": "AIX\n",
		"oslevel":  "7.3.0.0\n",
	})

	platform, family, version, err := PlatformInformationWithContext(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, "AIX", platform)
	assert.Equal(t, "AIX", family)
	assert.Equal(t, "7.3.0.0", version)
}

func TestKernelVersionWithContext(t *testing.T) {
	withMockInvoker(t, map[string]string{
		"oslevel -s": "7300-03-00-2446\n",
	})

	version, err := KernelVersionWithContext(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, "7300-03-00-2446", version)
}

func TestKernelArch(t *testing.T) {
	withMockInvoker(t, map[string]string{
		"bootinfo -y": "64\n",
	})

	arch, err := KernelArch()
	require.NoError(t, err)
	assert.Equal(t, "64", arch)
}

func TestVirtualizationWithContext(t *testing.T) {
	system, role, err := VirtualizationWithContext(context.TODO())
	require.NoError(t, err)
	// On a real AIX system, we expect either powervm or wpar
	if system != "" {
		assert.Contains(t, []string{"powervm", "wpar"}, system)
		assert.Equal(t, "guest", role)
	}
}

func TestVirtualizationWithContext_LPAR(t *testing.T) {
	withMockInvoker(t, map[string]string{
		"uname -W": "0\n",
		"uname -L": "25 soaix422\n",
	})

	system, role, err := VirtualizationWithContext(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, "powervm", system)
	assert.Equal(t, "guest", role)
}

func TestVirtualizationWithContext_WPAR(t *testing.T) {
	withMockInvoker(t, map[string]string{
		"uname -W": "2\n",
		"uname -L": "25 soaix422\n",
	})

	system, role, err := VirtualizationWithContext(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, "wpar", system)
	assert.Equal(t, "guest", role)
}

func TestVirtualizationWithContext_BareMetal(t *testing.T) {
	withMockInvoker(t, map[string]string{
		"uname -W": "0\n",
		"uname -L": "-1 NULL\n",
	})

	system, role, err := VirtualizationWithContext(context.TODO())
	require.NoError(t, err)
	assert.Empty(t, system)
	assert.Empty(t, role)
}
