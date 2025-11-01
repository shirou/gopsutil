// SPDX-License-Identifier: BSD-3-Clause
package cpu

import (
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestTimesEmpty(t *testing.T) {
	t.Setenv("HOST_PROC", "testdata/linux/times_empty")
	_, err := Times(true)
	require.NoErrorf(t, err, "Times(true) failed")
	_, err = Times(false)
	assert.NoErrorf(t, err, "Times(false) failed")
}

func TestParseStatLine_424(t *testing.T) {
	t.Setenv("HOST_PROC", "testdata/linux/424/proc")
	{
		l, err := Times(true)
		if err != nil || len(l) == 0 {
			t.Error("Times(true) failed")
		}
		t.Logf("Times(true): %#v", l)
	}
	{
		l, err := Times(false)
		if err != nil || len(l) == 0 {
			t.Error("Times(false) failed")
		}
		t.Logf("Times(false): %#v", l)
	}
}

func TestCountsAgainstLscpu(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), "lscpu")
	cmd.Env = []string{"LC_ALL=C"}
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			t.Skip("no lscpu to compare with")
		}
		t.Errorf("error executing lscpu: %v", err)
	}
	var threadsPerCore, coresPerSocket, sockets, books, drawers int
	books = 1
	drawers = 1
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "Thread(s) per core":
			threadsPerCore, _ = strconv.Atoi(strings.TrimSpace(fields[1]))
		case "Core(s) per socket":
			coresPerSocket, _ = strconv.Atoi(strings.TrimSpace(fields[1]))
		case "Socket(s)", "Socket(s) per book":
			sockets, _ = strconv.Atoi(strings.TrimSpace(fields[1]))
		case "Book(s) per drawer":
			books, _ = strconv.Atoi(strings.TrimSpace(fields[1]))
		case "Drawer(s)":
			drawers, _ = strconv.Atoi(strings.TrimSpace(fields[1]))
		}
	}
	if threadsPerCore == 0 || coresPerSocket == 0 || sockets == 0 {
		t.Errorf("missing info from lscpu: threadsPerCore=%d coresPerSocket=%d sockets=%d", threadsPerCore, coresPerSocket, sockets)
	}
	expectedPhysical := coresPerSocket * sockets * books * drawers
	expectedLogical := expectedPhysical * threadsPerCore
	physical, err := Counts(false)
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	logical, err := Counts(true)
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	assert.Equalf(t, expectedPhysical, physical, "expected %v, got %v", expectedPhysical, physical)
	assert.Equalf(t, expectedLogical, logical, "expected %v, got %v", expectedLogical, logical)
}

func TestCountsLogicalAndroid_1037(t *testing.T) { // https://github.com/shirou/gopsutil/issues/1037
	t.Setenv("HOST_PROC", "testdata/linux/1037/proc")

	count, err := Counts(true)
	require.NoError(t, err)
	expected := 8
	assert.Equalf(t, expected, count, "expected %v, got %v", expected, count)
}
