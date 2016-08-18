// +build darwin

package process

// #include <stdlib.h>
// #include <libproc.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// internal_GetProcessExe is a OS X specific way to get the full path to exe of a running process.
func internal_GetProcessExe(pid int) (string, error) {
	var c C.char // need a var for unsafe.Sizeof need a var
	const bufsize = C.PROC_PIDPATHINFO_MAXSIZE * unsafe.Sizeof(c)
	buffer := (*C.char)(C.malloc(C.size_t(bufsize)))
	defer C.free(unsafe.Pointer(buffer))

	ret, err := C.proc_pidpath(C.int(pid), unsafe.Pointer(buffer), C.uint32_t(bufsize))
	if err != nil {
		return "", err
	} else if ret <= 0 {
		return "", fmt.Errorf("Unknown error: proc_pidpath returned %d", ret)
	}

	gostr := C.GoString(buffer)
	return gostr, nil
}
