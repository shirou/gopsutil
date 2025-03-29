// SPDX-License-Identifier: BSD-3-Clause
package common

import (
	"context"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/common"
)

func TestReadlines(t *testing.T) {
	ret, err := ReadLines("common_test.go")
	require.NoError(t, err)
	assert.Containsf(t, ret[1], "package common", "could not read correctly")
}

func TestReadLinesOffsetN(t *testing.T) {
	ret, err := ReadLinesOffsetN("common_test.go", 3, 1)
	require.NoError(t, err)
	assert.Containsf(t, ret[0], `import (`, "could not read correctly")
}

func TestIntToString(t *testing.T) {
	assert.Equalf(t, "ABC", IntToString([]int8{65, 66, 67}), "could not convert")
}

func TestByteToString(t *testing.T) {
	assert.Equalf(t, "ABC", ByteToString([]byte{65, 66, 67}), "could not convert")

	assert.Equalf(t, "ABC", ByteToString([]byte{0, 65, 66, 67}), "could not convert")
}

func TestHexToUint32(t *testing.T) {
	assert.Equalf(t, uint32(4294967295), HexToUint32("FFFFFFFF"), "Could not convert")
}

func TestMustParseInt32(t *testing.T) {
	assert.Equalf(t, int32(11111), mustParseInt32("11111"), "could not parse")
}

func TestMustParseUint64(t *testing.T) {
	assert.Equalf(t, uint64(11111), mustParseUint64("11111"), "could not parse")
}

func TestMustParseFloat64(t *testing.T) {
	require.InDeltaf(t, float64(11111.11), mustParseFloat64("11111.11"), 0.01, "could not parse")
	require.InDeltaf(t, float64(11111), mustParseFloat64("11111"), 0.01, "could not parse")
}

func TestStringsContains(t *testing.T) {
	target, err := ReadLines("common_test.go")
	require.NoError(t, err)
	assert.Truef(t, StringsContains(target, "func TestStringsContains(t *testing.T) {"), "cloud not test correctly")
}

func TestPathExists(t *testing.T) {
	assert.Truef(t, PathExists("common_test.go"), "exists but return not exists")
	assert.Falsef(t, PathExists("should_not_exists.go"), "not exists but return exists")
}

func TestPathExistsWithContents(t *testing.T) {
	assert.Truef(t, PathExistsWithContents("common_test.go"), "exists but return not exists")
	assert.Falsef(t, PathExistsWithContents("should_not_exists.go"), "not exists but return exists")

	f, err := os.CreateTemp("", "empty_test.txt")
	require.NoErrorf(t, err, "CreateTemp failed, %s", err)
	defer os.Remove(f.Name()) // clean up

	assert.Falsef(t, PathExistsWithContents(f.Name()), "exists but no content file return true")
}

func TestHostEtc(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows doesn't have etc")
	}
	p := HostEtcWithContext(context.Background(), "mtab")
	assert.Equalf(t, "/etc/mtab", p, "invalid HostEtc, %s", p)
}

func TestGetSysctrlEnv(t *testing.T) {
	// Append case
	env := getSysctrlEnv([]string{"FOO=bar"})
	assert.Truef(t, reflect.DeepEqual(env, []string{"FOO=bar", "LC_ALL=C"}), "unexpected append result from getSysctrlEnv: %q", env)

	// Replace case
	env = getSysctrlEnv([]string{"FOO=bar", "LC_ALL=en_US.UTF-8"})
	assert.Truef(t, reflect.DeepEqual(env, []string{"FOO=bar", "LC_ALL=C"}), "unexpected replace result from getSysctrlEnv: %q", env)

	// Test against real env
	env = getSysctrlEnv(os.Environ())
	found := false
	for _, v := range env {
		if v == "LC_ALL=C" {
			found = true
			continue
		}
		require.Falsef(t, strings.HasPrefix(v, "LC_ALL"), "unexpected LC_ALL value: %q", v)
	}
	assert.Truef(t, found, "unexpected real result from getSysctrlEnv: %q", env)
}

func TestGetEnvDefault(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows doesn't have etc")
	}
	p := HostEtcWithContext(context.Background(), "mtab")
	assert.Equalf(t, "/etc/mtab", p, "invalid HostEtc, %s", p)
}

func TestGetEnvWithNoContext(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows doesn't have etc")
	}
	t.Setenv("HOST_ETC", "/bar")
	p := HostEtcWithContext(context.Background(), "mtab")
	assert.Equalf(t, "/bar/mtab", p, "invalid HostEtc, %s", p)
}

func TestGetEnvWithContextOverride(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows doesn't have etc")
	}
	t.Setenv("HOST_ETC", "/bar")
	ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{common.HostEtcEnvKey: "/foo"})
	p := HostEtcWithContext(ctx, "mtab")
	assert.Equalf(t, "/foo/mtab", p, "invalid HostEtc, %s", p)
}
