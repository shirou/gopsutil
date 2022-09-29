//go:build windows
// +build windows

package procstats

import (
	"unsafe"

	"github.com/shirou/gopsutil/v3/internal/common"
	"golang.org/x/sys/windows"
)

// represents a win32 thread status
// see <winternl.h> for possible values
type THREAD_STATE uint32

const (
	StateInitialized THREAD_STATE = iota
	StateReady
	StateRunning
	StateStandby
	StateTerminated
	StateWait
	StateTransition
	StateUnknown
)

func (s THREAD_STATE) String() string {
	switch s {
	case StateInitialized:
		return "StateInitialized"
	case StateReady:
		return "StateReady"
	case StateRunning:
		return "StateRunning"
	case StateStandby:
		return "StateStandby"
	case StateTerminated:
		return "StateTerminated"
	case StateWait:
		return "StateWait"
	case StateTransition:
		return "StateTransition"
	case StateUnknown:
		return "StateUnknown"
	default:
		return "<invalid state>"
	}
}

// "BUSY" thread means a thread that is either running, transitioning to run or on a state that demans resources
// see https://www.brendangregg.com/blog/2017-08-08/linux-load-averages.html
func (s THREAD_STATE) Busy() bool {
	switch s {
	case StateReady, StateRunning, StateStandby, StateTransition:
		return true
	default:
		return false
	}
}

// SYSTEM_THREAD_INFORMATION contains thread information as it is returned by NtQuerySystemInformation() API call
// look for its structure & documentation at:
// https://learn.microsoft.com/en-us/windows/win32/api/winternl/nf-winternl-ntquerysysteminformation
type SYSTEM_THREAD_INFORMATION struct {
	Reserved1     [3]int64
	Reserved2     uint32
	StartAddress  uintptr
	UniqueProcess windows.Handle
	UniqueThread  windows.Handle
	Priority      int32
	BasePriority  int32
	Reserved3     uint32
	ThreadState   THREAD_STATE
	WaitReason    uint32
}

// SYSTEM_PROCESS_INFORMATION is a convenience struct to have first thread address at hand
// for this technique to access to heterogeneous data, see:
// https://justen.codes/breaking-all-the-rules-using-go-to-call-windows-api-2cbfd8c79724
type SYSTEM_PROCESS_INFORMATION struct {
	windows.SYSTEM_PROCESS_INFORMATION
	ThreadsTable [1]SYSTEM_THREAD_INFORMATION
}

// Stats are the stats this package offers
type Stats struct {
	ProcessCount    uint32
	ThreadCount     uint32
	ThreadsByStatus map[THREAD_STATE]uint32
	Load            uint32 // number of threads that contribute to system load, see https://www.brendangregg.com/blog/2017-08-08/linux-load-averages.html
}

func EmptyStats() *Stats {
	return &Stats{
		ThreadsByStatus: make(map[THREAD_STATE]uint32),
	}
}

// AddProc increments process count and returns itself.
func (s *Stats) AddProc() *Stats {
	s.ProcessCount += 1
	return s
}

// AddThread increments thread count, also updates ThreadsByStatus based on the status,
// finally if the state represents a busy thread, it increments the load.
// returns the current stats structure pointer.
func (s *Stats) AddThread(state THREAD_STATE) *Stats {
	s.ThreadCount += 1
	s.ThreadsByStatus[state] += 1

	if state.Busy() {
		s.Load += 1
	}

	return s
}

// SystemProcessInformationWalk is a helper structure to walk through the raw bytes
// that NtQuerySystemInformation produces and get correct structures
type SystemProcessInformationWalk struct {
	SizeInBytes uint32 // buffer size
	Offset      uint32 // current offset
	Buffer      []byte // buffer with the data
}

// Process returns the process under current offset
func (w *SystemProcessInformationWalk) Process() *SYSTEM_PROCESS_INFORMATION {
	return (*SYSTEM_PROCESS_INFORMATION)(unsafe.Pointer(&w.Buffer[w.Offset]))
}

// Next moves offset to the next process structure
// it returns true if there are still more PENDING processess to iterate
// it returns false if there are no more PENDING processess to iterate
// calling Next() when there are no more processes, has no effect
func (w *SystemProcessInformationWalk) Next() bool {
	proc := w.Process()

	if proc.NextEntryOffset == 0 || proc.NextEntryOffset+w.Offset > w.SizeInBytes {
		return false // reached the end
	}

	w.Offset += proc.NextEntryOffset

	return true
}

// Stats calculate stats for all processes and their threads
func (w *SystemProcessInformationWalk) Stats() *Stats {
	stats := EmptyStats()

	for {
		proc := w.Process()

		stats.AddProc()

		WalkThreads(proc, func(t SYSTEM_THREAD_INFORMATION) {
			stats.AddThread(t.ThreadState)
		})

		if ok := w.Next(); !ok {
			break
		}
	}

	return stats
}

// WalkThreads() iterates over all threads of current process and applies given function
func WalkThreads(proc *SYSTEM_PROCESS_INFORMATION, fn func(t SYSTEM_THREAD_INFORMATION)) {
	for i := 0; i < int(proc.NumberOfThreads); i++ {
		thread := *(*SYSTEM_THREAD_INFORMATION)(unsafe.Pointer(
			uintptr(unsafe.Pointer(&proc.ThreadsTable[0])) +
				uintptr(i)*unsafe.Sizeof(proc.ThreadsTable[0]),
		))

		fn(thread)
	}
}

// GetSystemProcessInformation retrieves information of all procecess and threads
// see: https://learn.microsoft.com/en-us/windows/win32/api/winternl/nf-winternl-ntquerysysteminformation
// look for SystemProcessInformation and related structures SYSTEM_PROCESS_INFORMATION and SYSTEM_THREAD_INFORMATION
// the returned structure has methods to walk through the structure
func GetSystemProcessInformation() (*SystemProcessInformationWalk, error) {
	var (
		oneKb      uint32 = 1024
		allocKb    uint32 = 1
		allocBytes uint32 = allocKb * oneKb
		buffer     []byte
		usedBytes  uint32
	)

	buffer = make([]byte, allocBytes)

	st := common.CallWithExpandingBuffer(
		func() common.NtStatus {
			return common.NtQuerySystemInformation(
				windows.SystemProcessInformation,
				&buffer[0],
				allocBytes,
				&usedBytes,
			)
		},
		&buffer,
		&usedBytes,
	)

	if st.IsError() {
		return nil, st.Error()
	}

	return &SystemProcessInformationWalk{
		SizeInBytes: usedBytes,
		Offset:      0,
		Buffer:      buffer,
	}, nil
}
