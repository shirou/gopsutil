// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package host

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/common"
)

type fakeLoginctlInvoke struct {
	sessions       string
	sessionOutputs map[string]string
}

func (f fakeLoginctlInvoke) Command(name string, arg ...string) ([]byte, error) {
	return f.CommandWithContext(context.Background(), name, arg...)
}

func (f fakeLoginctlInvoke) CommandWithContext(_ context.Context, name string, arg ...string) ([]byte, error) {
	if name != "loginctl" || len(arg) == 0 {
		return nil, fmt.Errorf("unexpected command: %s %v", name, arg)
	}
	switch arg[0] {
	case "list-sessions":
		if !slices.Equal(arg, []string{"list-sessions", "--no-legend", "--no-pager"}) {
			return nil, fmt.Errorf("unexpected list-sessions args: %v", arg)
		}
		return []byte(f.sessions), nil
	case "show-session":
		want := []string{"-p", "Name", "-p", "TTY", "-p", "RemoteHost", "-p", "Timestamp", "-p", "Class", "-p", "Seat"}
		if len(arg) != 2+len(want) || !slices.Equal(arg[2:], want) {
			return nil, fmt.Errorf("unexpected show-session args: %v", arg)
		}
		out, ok := f.sessionOutputs[arg[1]]
		if !ok {
			return nil, fmt.Errorf("unexpected session id: %s", arg[1])
		}
		return []byte(out), nil
	default:
		return nil, fmt.Errorf("unexpected subcommand: %s", arg[0])
	}
}

func useFakeLoginctl(t *testing.T, fake fakeLoginctlInvoke) {
	t.Helper()
	old := invoke
	invoke = fake
	t.Cleanup(func() { invoke = old })
}

func useUTC(t *testing.T) {
	t.Helper()
	old := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = old })
}

func TestUsersFromLoginctl(t *testing.T) {
	useUTC(t)
	// Right-aligned session ids, as printed by loginctl --no-legend.
	useFakeLoginctl(t, fakeLoginctlInvoke{
		sessions: "  2771     0 root -    4242 user pts/1 no -\n",
		sessionOutputs: map[string]string{
			"2771": "Name=root\nTTY=pts/1\nRemoteHost=10.5.22.31\nTimestamp=Thu 2026-01-22 14:51:57 UTC\nClass=user\nSeat=\n",
		},
	})

	got, err := usersFromLoginctlWithContext(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "root", got[0].User)
	assert.Equal(t, "pts/1", got[0].Terminal)
	assert.Equal(t, "10.5.22.31", got[0].Host)
	assert.Equal(t, 1769093517, got[0].Started)
}

// Sessions as listed on Ubuntu 26.04 (systemd 259): a login plus the per-user
// manager, and on a desktop the greeter and a graphical login (seat, no tty).
func TestUsersFromLoginctlFiltersSessions(t *testing.T) {
	useUTC(t)
	useFakeLoginctl(t, fakeLoginctlInvoke{
		sessions: "   5158 10049 shirou -    961981 user    - no -\n" +
			"   5159 10049 shirou -    961996 manager - no -\n" +
			"     c1   120 gdm      seat0 1200   greeter tty1 no -\n" +
			"     c2  1000 alice    seat0 1500   user    - no -\n" +
			"     c3  1000 alice    -     1600   user-early - no -\n",
		sessionOutputs: map[string]string{
			"5158": "Name=shirou\nTTY=pts/0\nRemoteHost=\nTimestamp=Mon 2026-10-05 08:00:00 UTC\nClass=user\nSeat=\n",
			"5159": "Name=shirou\nTTY=\nRemoteHost=\nTimestamp=Mon 2026-10-05 08:00:00 UTC\nClass=manager\nSeat=\n",
			"c1":   "Name=gdm\nTTY=tty1\nRemoteHost=\nTimestamp=Mon 2026-10-05 08:00:00 UTC\nClass=greeter\nSeat=seat0\n",
			"c2":   "Name=alice\nTTY=\nRemoteHost=\nTimestamp=Mon 2026-10-05 08:01:00 UTC\nClass=user\nSeat=seat0\n",
			"c3":   "Name=alice\nTTY=\nRemoteHost=\nTimestamp=Mon 2026-10-05 08:01:00 UTC\nClass=user-early\nSeat=\n",
		},
	})

	got, err := usersFromLoginctlWithContext(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "shirou", got[0].User)
	assert.Equal(t, "pts/0", got[0].Terminal)
	assert.Equal(t, 1791187200, got[0].Started)
	// Graphical login: only a seat, no tty; the seat is reported as terminal.
	assert.Equal(t, "alice", got[1].User)
	assert.Equal(t, "seat0", got[1].Terminal)
}

func TestUsersFromLoginctlTimestampOffsetZone(t *testing.T) {
	old := time.Local
	time.Local = time.FixedZone("", 3*3600+1800)
	defer func() { time.Local = old }()
	useFakeLoginctl(t, fakeLoginctlInvoke{
		sessions: "1 0 root - 1 user pts/0 no -\n",
		sessionOutputs: map[string]string{
			// "+0330" is rejected by time.Parse's MST layout.
			"1": "Name=root\nTTY=pts/0\nRemoteHost=\nTimestamp=Mon 2026-10-05 11:30:00 +0330\nClass=user\nSeat=\n",
		},
	})

	got, err := usersFromLoginctlWithContext(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1791187200, got[0].Started)
}

func TestUsersFromLoginctlNoSessions(t *testing.T) {
	useFakeLoginctl(t, fakeLoginctlInvoke{})

	got, err := usersFromLoginctlWithContext(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestUsersFromLoginctlAllSessionsFail(t *testing.T) {
	useFakeLoginctl(t, fakeLoginctlInvoke{sessions: "1 0 root - 1 user pts/0 no -\n"})

	_, err := usersFromLoginctlWithContext(context.Background())
	require.Error(t, err)
}

// Users falls back to loginctl when the utmp file does not exist.
func TestUsersFallsBackToLoginctlWithoutUtmp(t *testing.T) {
	useFakeLoginctl(t, fakeLoginctlInvoke{
		sessions: "1 0 root - 1 user pts/0 no -\n",
		sessionOutputs: map[string]string{
			"1": "Name=root\nTTY=pts/0\nRemoteHost=\nTimestamp=Mon 2026-10-05 08:00:00 UTC\nClass=user\nSeat=\n",
		},
	})
	ctx := context.WithValue(context.Background(),
		common.EnvKey,
		common.EnvMap{common.HostVarEnvKey: t.TempDir()},
	)

	got, err := UsersWithContext(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "root", got[0].User)
}

// When utmp is missing and loginctl fails too, the original utmp error is
// returned so that os.IsNotExist keeps working.
func TestUsersWithoutUtmpAndLoginctlKeepsNotExist(t *testing.T) {
	useFakeLoginctl(t, fakeLoginctlInvoke{sessions: "1 0 root - 1 user pts/0 no -\n"})
	ctx := context.WithValue(context.Background(),
		common.EnvKey,
		common.EnvMap{common.HostVarEnvKey: t.TempDir()},
	)

	_, err := UsersWithContext(ctx)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err), "got %v", err)
}

func TestGetRedhatishVersion(t *testing.T) {
	var ret string
	c := []string{"Rawhide"}
	ret = getRedhatishVersion(c)
	assert.Equalf(t, "rawhide", ret, "Could not get version rawhide: %v", ret)

	c = []string{"Fedora release 15 (Lovelock)"}
	ret = getRedhatishVersion(c)
	assert.Equalf(t, "15", ret, "Could not get version fedora: %v", ret)

	c = []string{"Enterprise Linux Server release 5.5 (Carthage)"}
	ret = getRedhatishVersion(c)
	assert.Equalf(t, "5.5", ret, "Could not get version redhat enterprise: %v", ret)

	c = []string{""}
	ret = getRedhatishVersion(c)
	assert.Emptyf(t, ret, "Could not get version with no value: %v", ret)
}

func TestGetRedhatishPlatform(t *testing.T) {
	var ret string
	c := []string{"red hat"}
	ret = getRedhatishPlatform(c)
	assert.Equalf(t, "redhat", ret, "Could not get platform redhat: %v", ret)

	c = []string{"Fedora release 15 (Lovelock)"}
	ret = getRedhatishPlatform(c)
	assert.Equalf(t, "fedora", ret, "Could not get platform fedora: %v", ret)

	c = []string{"Enterprise Linux Server release 5.5 (Carthage)"}
	ret = getRedhatishPlatform(c)
	assert.Equalf(t, "enterprise", ret, "Could not get platform redhat enterprise: %v", ret)

	c = []string{""}
	ret = getRedhatishPlatform(c)
	assert.Emptyf(t, ret, "Could not get platform with no value: %v", ret)
}

func TestGetlsbStruct(t *testing.T) {
	cases := []struct {
		root        string
		id          string
		release     string
		codename    string
		description string
	}{
		{"arch", "Arch", "rolling", "", "Arch Linux"},
		{"ubuntu_22_04", "Ubuntu", "22.04", "jammy", "Ubuntu 22.04.2 LTS"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.root, func(t *testing.T) {
			ctx := context.WithValue(context.Background(),
				common.EnvKey,
				common.EnvMap{common.HostEtcEnvKey: "./testdata/linux/lsbStruct/" + tt.root},
			)

			v, err := getlsbStruct(ctx)
			require.NoError(t, err)
			assert.Equalf(t, v.ID, tt.id, "ID: want %v, got %v", tt.id, v.ID)
			assert.Equalf(t, v.Release, tt.release, "Release: want %v, got %v", tt.release, v.Release)
			assert.Equalf(t, v.Codename, tt.codename, "Codename: want %v, got %v", tt.codename, v.Codename)
			assert.Equalf(t, v.Description, tt.description, "Description: want %v, got %v", tt.description, v.Description)

			t.Log(v)
		})
	}
}
