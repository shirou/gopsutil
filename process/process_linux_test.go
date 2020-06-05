// +build linux

package process

import (
	"os"
	"testing"

	log "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

var sink map[int32]*FilledProcess

func BenchmarkLinuxAllProcessesOnPostgresProcFS(b *testing.B) {
	var err error
	errCount := 0

	// Set procfs to testdata location, does not include:
	//  /fd  - fd count
	//  /exe - symlink
	//  /cwd - symlink
	os.Setenv("HOST_PROC", "resources/linux_postgres/proc")
	defer os.Unsetenv("HOST_PROC")

	// Disable logging (as it'll be noisy)
	log.ReplaceLogger(log.Disabled)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if sink, err = AllProcesses(); err != nil {
			errCount++
		}
	}

	if errCount > 0 {
		b.Fatal("non-zero errors for AllProcesses", errCount)
	}
}

func BenchmarkLinuxAllProcessesOnLocalProcFS(b *testing.B) {
	var err error
	errCount := 0

	// Disable logging (as it'll be noisy)
	log.ReplaceLogger(log.Disabled)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if sink, err = AllProcesses(); err != nil {
			errCount++
		}
	}

	if errCount > 0 {
		b.Fatal("non-zero errors for AllProcesses", errCount)
	}
}

func TestFillNSPidFromStatus(t *testing.T) {
	err := os.Setenv("HOST_PROC", "./resources/linux_inner_ns/proc")
	defer os.Unsetenv("HOST_PROC")
	assert.Nil(t, err)

	// Process spawned from the host
	// NSpid:	6320
	p1, err := NewProcess(6320)
	assert.Nil(t, err)
	p1.fillFromStatus()
	assert.Equal(t, int32(6320), p1.NSPid)

	// Process spawned from P1 with its own namespace
	// NSpid:	6321	1
	p2, err := NewProcess(6321)
	assert.Nil(t, err)
	p2.fillFromStatus()
	assert.Equal(t, int32(1), p2.NSPid)

	// Process spawned from P2 with its own namespace
	// NSpid:	6322	2	1
	p3, err := NewProcess(6322)
	assert.Nil(t, err)
	p3.fillFromStatus()
	assert.Equal(t, int32(1), p3.NSPid)
}
