// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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
	withMockInvoker(t, map[string]string{
		"who": "root        pts/0       Feb 27 06:58     (192.168.1.1)\nadmin       pts/1       Feb 27 07:30\n",
	})

	users, err := UsersWithContext(context.TODO())
	require.NoError(t, err)
	require.Len(t, users, 2)

	assert.Equal(t, "root", users[0].User)
	assert.Equal(t, "pts/0", users[0].Terminal)
	assert.Equal(t, "192.168.1.1", users[0].Host)
	assert.NotZero(t, users[0].Started)

	assert.Equal(t, "admin", users[1].User)
	assert.Equal(t, "pts/1", users[1].Terminal)
	assert.Empty(t, users[1].Host)
	assert.NotZero(t, users[1].Started)
}

func TestUsersWithContextEmpty(t *testing.T) {
	withMockInvoker(t, map[string]string{
		"who": "",
	})

	users, err := UsersWithContext(context.TODO())
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestParseWhoTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		month    string
		day      string
		hhmm     string
		now      time.Time
		wantZero bool
	}{
		{
			name:  "normal case",
			month: "Feb", day: "27", hhmm: "06:58",
			now: time.Date(2026, time.February, 27, 12, 0, 0, 0, time.Local),
		},
		{
			name:  "year boundary - Dec login, Jan now",
			month: "Dec", day: "31", hhmm: "23:50",
			now: time.Date(2026, time.January, 1, 0, 5, 0, 0, time.Local),
		},
		{
			name:  "same month earlier day",
			month: "Jan", day: "15", hhmm: "10:00",
			now: time.Date(2026, time.January, 20, 12, 0, 0, 0, time.Local),
		},
		{
			name:  "invalid month",
			month: "Xyz", day: "99", hhmm: "25:61",
			now:      time.Now(),
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseWhoTimestamp(tt.month, tt.day, tt.hhmm, tt.now)
			if tt.wantZero {
				assert.Zero(t, result)
				return
			}
			assert.Positive(t, result)

			ts := time.Unix(int64(result), 0)
			if tt.month == "Dec" && tt.now.Month() == time.January {
				assert.Equal(t, tt.now.Year()-1, ts.Year(), "year boundary: should use previous year")
			} else {
				assert.Equal(t, tt.now.Year(), ts.Year())
			}
		})
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
