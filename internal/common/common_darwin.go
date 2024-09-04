// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package common

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

func DoSysctrlWithContext(ctx context.Context, mib string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "sysctl", "-n", mib)
	cmd.Env = getSysctrlEnv(os.Environ())
	out, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	v := strings.Replace(string(out), "{ ", "", 1)
	v = strings.Replace(string(v), " }", "", 1)
	values := strings.Fields(string(v))

	return values, nil
}

func CallSyscall(mib []int32) ([]byte, uint64, error) {
	miblen := uint64(len(mib))

	// get required buffer size
	length := uint64(0)
	_, _, err := unix.Syscall6(
		202, // unix.SYS___SYSCTL https://github.com/golang/sys/blob/76b94024e4b621e672466e8db3d7f084e7ddcad2/unix/zsysnum_darwin_amd64.go#L146
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(miblen),
		0,
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if err != 0 {
		var b []byte
		return b, length, err
	}
	if length == 0 {
		var b []byte
		return b, length, err
	}
	// get proc info itself
	buf := make([]byte, length)
	_, _, err = unix.Syscall6(
		202, // unix.SYS___SYSCTL https://github.com/golang/sys/blob/76b94024e4b621e672466e8db3d7f084e7ddcad2/unix/zsysnum_darwin_amd64.go#L146
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(miblen),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if err != 0 {
		return buf, length, err
	}

	return buf, length, nil
}

// Library represents a dynamic library loaded by purego.
type Library struct {
	addr  uintptr
	path  string
	close func()
}

// library paths
const (
	IOKit          = "/System/Library/Frameworks/IOKit.framework/IOKit"
	CoreFoundation = "/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation"
	Kernel         = "/usr/lib/system/libsystem_kernel.dylib"
)

func NewLibrary(path string) (*Library, error) {
	lib, err := purego.Dlopen(path, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, err
	}

	closeFunc := func() {
		purego.Dlclose(lib)
	}

	return &Library{
		addr:  lib,
		path:  path,
		close: closeFunc,
	}, nil
}

func (lib *Library) Dlsym(symbol string) (uintptr, error) {
	return purego.Dlsym(lib.addr, symbol)
}

func GetFunc[T any](lib *Library, symbol string) T {
	var fptr T
	purego.RegisterLibFunc(&fptr, lib.addr, symbol)
	return fptr
}

func (lib *Library) Close() {
	lib.close()
}

// status codes
const (
	KERN_SUCCESS = 0
)

// IOKit functions and symbols.
type (
	IOServiceGetMatchingServiceFunc func(mainPort uint32, matching uintptr) uint32
	IOServiceMatchingFunc           func(name string) unsafe.Pointer
	IOServiceOpenFunc               func(service, owningTask, connType uint32, connect *uint32) int
	IOServiceCloseFunc              func(connect uint32) int
	IOObjectReleaseFunc             func(object uint32) int
	IOConnectCallStructMethodFunc   func(connection, selector uint32, inputStruct, inputStructCnt, outputStruct uintptr, outputStructCnt *uintptr) int

	IOHIDEventSystemClientCreateFunc      func(allocator uintptr) unsafe.Pointer
	IOHIDEventSystemClientSetMatchingFunc func(client, match uintptr) int
	IOHIDServiceClientCopyEventFunc       func(service uintptr, eventType int64,
		options int32, timeout int64) unsafe.Pointer
	IOHIDServiceClientCopyPropertyFunc     func(service, property uintptr) unsafe.Pointer
	IOHIDEventGetFloatValueFunc            func(event uintptr, field int32) float64
	IOHIDEventSystemClientCopyServicesFunc func(client uintptr) unsafe.Pointer
)

const (
	IOServiceGetMatchingServiceSym = "IOServiceGetMatchingService"
	IOServiceMatchingSym           = "IOServiceMatching"
	IOServiceOpenSym               = "IOServiceOpen"
	IOServiceCloseSym              = "IOServiceClose"
	IOObjectReleaseSym             = "IOObjectRelease"
	IOConnectCallStructMethodSym   = "IOConnectCallStructMethod"

	IOHIDEventSystemClientCreateSym       = "IOHIDEventSystemClientCreate"
	IOHIDEventSystemClientSetMatchingSym  = "IOHIDEventSystemClientSetMatching"
	IOHIDServiceClientCopyEventSym        = "IOHIDServiceClientCopyEvent"
	IOHIDServiceClientCopyPropertySym     = "IOHIDServiceClientCopyProperty"
	IOHIDEventGetFloatValueSym            = "IOHIDEventGetFloatValue"
	IOHIDEventSystemClientCopyServicesSym = "IOHIDEventSystemClientCopyServices"
)

const (
	KIOHIDEventTypeTemperature = 15
)

// CoreFoundation functions and symbols.
type (
	CFNumberCreateFunc     func(allocator uintptr, theType int32, valuePtr uintptr) unsafe.Pointer
	CFDictionaryCreateFunc func(allocator uintptr, keys, values *unsafe.Pointer, numValues int32,
		keyCallBacks, valueCallBacks uintptr) unsafe.Pointer
	CFArrayGetCountFunc           func(theArray uintptr) int32
	CFArrayGetValueAtIndexFunc    func(theArray uintptr, index int32) unsafe.Pointer
	CFStringCreateMutableFunc     func(alloc uintptr, maxLength int32) unsafe.Pointer
	CFStringGetLengthFunc         func(theString uintptr) int32
	CFStringGetCStringFunc        func(theString uintptr, buffer *byte, bufferSize int32, encoding uint32)
	CFStringCreateWithCStringFunc func(alloc uintptr, cStr string, encoding uint32) unsafe.Pointer
	CFReleaseFunc                 func(cf uintptr)
)

const (
	CFNumberCreateSym            = "CFNumberCreate"
	CFDictionaryCreateSym        = "CFDictionaryCreate"
	CFArrayGetCountSym           = "CFArrayGetCount"
	CFArrayGetValueAtIndexSym    = "CFArrayGetValueAtIndex"
	CFStringCreateMutableSym     = "CFStringCreateMutable"
	CFStringGetLengthSym         = "CFStringGetLength"
	CFStringGetCStringSym        = "CFStringGetCString"
	CFStringCreateWithCStringSym = "CFStringCreateWithCString"
	CFReleaseSym                 = "CFRelease"
)

const (
	KCFStringEncodingUTF8 = 0x08000100
	KCFNumberIntType      = 9
	KCFAllocatorDefault   = 0
)

// Kernel functions and symbols.
type (
	HostProcessorInfoFunc func(host uint32, flavor int, outProcessorCount *uint32, outProcessorInfo uintptr,
		outProcessorInfoCnt *uint32) int
	HostStatisticsFunc func(host uint32, flavor int, hostInfoOut uintptr, hostInfoOutCnt *uint32) int
	MachHostSelfFunc   func() uint32
	MachTaskSelfFunc   func() uint32
	VMDeallocateFunc   func(targetTask uint32, vmAddress, vmSize uintptr) int
)

const (
	HostProcessorInfoSym = "host_processor_info"
	HostStatisticsSym    = "host_statistics"
	MachHostSelfSym      = "mach_host_self"
	MachTaskSelfSym      = "mach_task_self"
	VMDeallocateSym      = "vm_deallocate"
)

const (
	HOST_VM_INFO       = 2
	HOST_CPU_LOAD_INFO = 3

	HOST_VM_INFO_COUNT = 0xf
)

// SMC represents a SMC instance.
type SMC struct {
	lib        *Library
	conn       uint32
	callStruct IOConnectCallStructMethodFunc
}

const ioServiceSMC = "AppleSMC"

const (
	KSMCUserClientOpen  = 0
	KSMCUserClientClose = 1
	KSMCHandleYPCEvent  = 2
	KSMCReadKey         = 5
	KSMCWriteKey        = 6
	KSMCGetKeyCount     = 7
	KSMCGetKeyFromIndex = 8
	KSMCGetKeyInfo      = 9
)

const (
	KSMCSuccess     = 0
	KSMCError       = 1
	KSMCKeyNotFound = 132
)

func NewSMC(ioKit *Library) (*SMC, error) {
	if ioKit.path != IOKit {
		return nil, fmt.Errorf("library is not IOKit")
	}

	ioServiceGetMatchingService := GetFunc[IOServiceGetMatchingServiceFunc](ioKit, IOServiceGetMatchingServiceSym)
	ioServiceMatching := GetFunc[IOServiceMatchingFunc](ioKit, IOServiceMatchingSym)
	ioServiceOpen := GetFunc[IOServiceOpenFunc](ioKit, IOServiceOpenSym)
	ioObjectRelease := GetFunc[IOObjectReleaseFunc](ioKit, IOObjectReleaseSym)
	machTaskSelf := GetFunc[MachTaskSelfFunc](ioKit, MachTaskSelfSym)

	ioConnectCallStructMethod := GetFunc[IOConnectCallStructMethodFunc](ioKit, IOConnectCallStructMethodSym)

	service := ioServiceGetMatchingService(0, uintptr(ioServiceMatching(ioServiceSMC)))
	if service == 0 {
		return nil, fmt.Errorf("ERROR: %s NOT FOUND", ioServiceSMC)
	}

	var conn uint32
	if result := ioServiceOpen(service, machTaskSelf(), 0, &conn); result != 0 {
		return nil, fmt.Errorf("ERROR: IOServiceOpen failed")
	}

	ioObjectRelease(service)
	return &SMC{
		lib:        ioKit,
		conn:       conn,
		callStruct: ioConnectCallStructMethod,
	}, nil
}

func (s *SMC) CallStruct(selector uint32, inputStruct, inputStructCnt, outputStruct uintptr, outputStructCnt *uintptr) int {
	return s.callStruct(s.conn, selector, inputStruct, inputStructCnt, outputStruct, outputStructCnt)
}

func (s *SMC) Close() error {
	ioServiceClose := GetFunc[IOServiceCloseFunc](s.lib, IOServiceCloseSym)

	if result := ioServiceClose(s.conn); result != 0 {
		return fmt.Errorf("ERROR: IOServiceClose failed")
	}
	return nil
}
