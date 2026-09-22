// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package host

import (
	"context"
	"fmt"
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
		return []byte(f.sessions), nil
	case "show-session":
		out, ok := f.sessionOutputs[arg[1]]
		if !ok {
			return nil, fmt.Errorf("unexpected session id: %s", arg[1])
		}
		return []byte(out), nil
	default:
		return nil, fmt.Errorf("unexpected subcommand: %s", arg[0])
	}
}

func TestUsersFromLoginctl(t *testing.T) {
	fake := fakeLoginctlInvoke{
		sessions: `[{"session":"2771","uid":0,"user":"root","seat":null,"tty":null,"state":"closing","idle":false,"since":null}]`,
		sessionOutputs: map[string]string{
			"2771": "Name=root\nTTY=pts/1\nRemoteHost=10.5.22.31\nTimestamp=Thu 2026-01-22 14:51:57 CET\n",
		},
	}

	old := invoke
	invoke = fake
	defer func() { invoke = old }()

	got, err := usersFromLoginctlWithContext(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "root", got[0].User)
	assert.Equal(t, "pts/1", got[0].Terminal)
	assert.Equal(t, "10.5.22.31", got[0].Host)

	wantTime, err := time.Parse("Mon 2006-01-02 15:04:05 MST", "Thu 2026-01-22 14:51:57 CET")
	require.NoError(t, err)
	assert.Equal(t, int(wantTime.Unix()), got[0].Started)
}

func TestUsersFromLoginctlNoSessions(t *testing.T) {
	fake := fakeLoginctlInvoke{sessions: `[]`}

	old := invoke
	invoke = fake
	defer func() { invoke = old }()

	got, err := usersFromLoginctlWithContext(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
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
