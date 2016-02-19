// +build darwin

package mem

/*
#include <unistd.h>
#include <mach/mach_host.h>
*/
import "C"

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"

	"github.com/shirou/gopsutil/internal/common"
)

func getPageSize() (uint64, error) {
	out, err := exec.Command("pagesize").Output()
	if err != nil {
		return 0, err
	}
	o := strings.TrimSpace(string(out))
	p, err := strconv.ParseUint(o, 10, 64)
	if err != nil {
		return 0, err
	}
	return p, nil
}

// VirtualMemory returns VirtualmemoryStat.
func VirtualMemory() (*VirtualMemoryStat, error) {
	count := C.mach_msg_type_number_t(C.HOST_VM_INFO_COUNT)
	var vmstat C.vm_statistics_data_t

	status := C.host_statistics(C.host_t(C.mach_host_self()),
		C.HOST_VM_INFO,
		C.host_info_t(unsafe.Pointer(&vmstat)),
		&count)

	if status != C.KERN_SUCCESS {
		return nil, fmt.Errorf("host_statistics error=%d", status)
	}

	totalCount := vmstat.wire_count +
		vmstat.active_count +
		vmstat.inactive_count +
		vmstat.free_count

	availableCount := vmstat.inactive_count + vmstat.free_count
	usedPercent := 100 * float64(totalCount-availableCount) / float64(totalCount)

	usedCount := totalCount - vmstat.free_count

	pageSize := uint64(C.getpagesize())
	return &VirtualMemoryStat{
		Total:       pageSize * uint64(totalCount),
		Available:   pageSize * uint64(availableCount),
		Used:        pageSize * uint64(usedCount),
		UsedPercent: usedPercent,
		Free:        pageSize * uint64(vmstat.free_count),
		Active:      pageSize * uint64(vmstat.active_count),
		Inactive:    pageSize * uint64(vmstat.inactive_count),
		Wired:       pageSize * uint64(vmstat.wire_count),
	}, nil
}

// SwapMemory returns swapinfo.
func SwapMemory() (*SwapMemoryStat, error) {
	var ret *SwapMemoryStat

	swapUsage, err := common.DoSysctrl("vm.swapusage")
	if err != nil {
		return ret, err
	}

	total := strings.Replace(swapUsage[2], "M", "", 1)
	used := strings.Replace(swapUsage[5], "M", "", 1)
	free := strings.Replace(swapUsage[8], "M", "", 1)

	total_v, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return nil, err
	}
	used_v, err := strconv.ParseFloat(used, 64)
	if err != nil {
		return nil, err
	}
	free_v, err := strconv.ParseFloat(free, 64)
	if err != nil {
		return nil, err
	}

	u := float64(0)
	if total_v != 0 {
		u = ((total_v - free_v) / total_v) * 100.0
	}

	// vm.swapusage shows "M", multiply 1000
	ret = &SwapMemoryStat{
		Total:       uint64(total_v * 1000),
		Used:        uint64(used_v * 1000),
		Free:        uint64(free_v * 1000),
		UsedPercent: u,
	}

	return ret, nil
}
