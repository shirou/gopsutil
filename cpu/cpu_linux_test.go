package cpu

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimesEmpty(t *testing.T) {
	t.Setenv("HOST_PROC", "testdata/linux/times_empty")
	_, err := Times(true)
	if err != nil {
		t.Error("Times(true) failed")
	}
	_, err = Times(false)
	if err != nil {
		t.Error("Times(false) failed")
	}
}

func TestCPUparseStatLine_424(t *testing.T) {
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

func TestCPUCountsAgainstLscpu(t *testing.T) {
	cmd := exec.Command("lscpu")
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
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	logical, err := Counts(true)
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if expectedPhysical != physical {
		t.Errorf("expected %v, got %v", expectedPhysical, physical)
	}
	if expectedLogical != logical {
		t.Errorf("expected %v, got %v", expectedLogical, logical)
	}
}

func TestCPUCountsLogicalAndroid_1037(t *testing.T) { // https://github.com/shirou/gopsutil/issues/1037
	t.Setenv("HOST_PROC", "testdata/linux/1037/proc")

	count, err := Counts(true)
	if err != nil {
		t.Errorf("error %v", err)
	}
	expected := 8
	if count != expected {
		t.Errorf("expected %v, got %v", expected, count)
	}
}

func TestCPUPressure(t *testing.T) {
	assert := assert.New(t)
	t.Setenv("HOST_PROC", "testdata/linux/proc")
	s, err := CPUPressure()
	assert.NoError(err)
	assert.Equal(1.29, s.SomeAvg10)
	assert.Equal(1.38, s.SomeAvg60)
	assert.Equal(1.39, s.SomeAvg300)
	assert.Equal(uint64(3825360042), s.SomeTotal)
	assert.Equal(0.00, s.FullAvg10)
	assert.Equal(0.00, s.FullAvg60)
	assert.Equal(0.00, s.FullAvg300)
	assert.Equal(uint64(0), s.FullTotal)
}
