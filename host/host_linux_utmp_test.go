// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package host

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/common"
)

func TestUsersWithContextUtmp(t *testing.T) {
	endian := "le"
	if binary.NativeEndian.Uint16([]byte{0, 1}) == 1 {
		endian = "be"
	}
	fixture := fmt.Sprintf("utmp_%d_%s", sizeOfUtmp, endian)

	data, err := os.ReadFile(filepath.Join("testdata", "linux", "utmp", fixture))
	require.NoError(t, err)

	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "run"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "run", "utmp"), data, 0o600))
	ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{
		common.HostVarEnvKey: root,
	})

	users, err := UsersWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, []UserStat{{
		User:     "audit-user",
		Terminal: "pts/42",
		Host:     "example.test",
		Started:  1700000000,
	}}, users)
}
