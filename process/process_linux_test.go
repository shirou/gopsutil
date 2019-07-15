// +build linux

package process

import (
	"os"
	"testing"

	log "github.com/cihub/seelog"
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
