// +build darwin

package cpu

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	common "github.com/shirou/gopsutil/common"
)

var HELPER_PATH = filepath.Join(os.Getenv("HOME"), ".gopsutil_cpu_helper")

// enable cpu helper. It may become security problem.
// This env valiable approach will be changed.
const HELPER_ENABLE_ENV = "ALLLOW_INSECURE_CPU_HELPER"

var ClocksPerSec = float64(100)

func init() {
	out, err := exec.Command("/usr/bin/getconf", "CLK_TCK").Output()
	// ignore errors
	if err == nil {
		i, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err == nil {
			ClocksPerSec = float64(i)
		}
	}

	// adhoc compile on the host. Errors will be ignored.
	// gcc is required to compile.
	if !common.PathExists(HELPER_PATH) && os.Getenv(HELPER_ENABLE_ENV) == "yes" {
		cmd := exec.Command("gcc", "-o", HELPER_PATH, "-x", "c", "-")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return
		}
		io.WriteString(stdin, cpu_helper_src)
		stdin.Close()
		cmd.Output()
	}
}

func CPUTimes(percpu bool) ([]CPUTimesStat, error) {
	var ret []CPUTimesStat
	if !common.PathExists(HELPER_PATH) {
		return nil, fmt.Errorf("could not get cpu time")
	}

	out, err := exec.Command(HELPER_PATH).Output()
	if err != nil {
		return ret, err
	}

	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Split(string(line), ",")
		if len(f) != 5 {
			continue
		}
		cpu, err := strconv.ParseFloat(f[0], 64)
		if err != nil {
			return ret, err
		}
		// cpu:99 means total, so just ignore if percpu
		if (percpu && cpu == 99) || (!percpu && cpu != 99) {
			continue
		}
		user, err := strconv.ParseFloat(f[1], 64)
		if err != nil {
			return ret, err
		}
		sys, err := strconv.ParseFloat(f[2], 64)
		if err != nil {
			return ret, err
		}
		idle, err := strconv.ParseFloat(f[3], 64)
		if err != nil {
			return ret, err
		}
		nice, err := strconv.ParseFloat(f[4], 64)
		if err != nil {
			return ret, err
		}
		c := CPUTimesStat{
			User:   float64(user / ClocksPerSec),
			Nice:   float64(nice / ClocksPerSec),
			System: float64(sys / ClocksPerSec),
			Idle:   float64(idle / ClocksPerSec),
		}
		if !percpu {
			c.CPU = "cpu-total"
		} else {
			c.CPU = fmt.Sprintf("cpu%d", uint16(cpu))
		}
		ret = append(ret, c)
	}
	return ret, nil
}

// Returns only one CPUInfoStat on FreeBSD
func CPUInfo() ([]CPUInfoStat, error) {
	var ret []CPUInfoStat

	out, err := exec.Command("/usr/sbin/sysctl", "machdep.cpu").Output()
	if err != nil {
		return ret, err
	}

	c := CPUInfoStat{}
	for _, line := range strings.Split(string(out), "\n") {
		values := strings.Fields(line)
		if len(values) < 1 {
			continue
		}

		t, err := strconv.ParseInt(values[1], 10, 64)
		// err is not checked here because some value is string.
		if strings.HasPrefix(line, "machdep.cpu.brand_string") {
			c.ModelName = strings.Join(values[1:], " ")
		} else if strings.HasPrefix(line, "machdep.cpu.family") {
			c.Family = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.model") {
			c.Model = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.stepping") {
			if err != nil {
				return ret, err
			}
			c.Stepping = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.leaf7_features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.extfeatures") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.core_count") {
			if err != nil {
				return ret, err
			}
			c.Cores = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.cache.size") {
			if err != nil {
				return ret, err
			}
			c.CacheSize = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.vendor") {
			c.VendorID = values[1]
		}

		// TODO:
		// c.Mhz = mustParseFloat64(values[1])
	}

	return append(ret, c), nil
}

const cpu_helper_src = `
#include <mach/mach.h>
#include <mach/mach_error.h>
#include <stdio.h>

int main() {
  	natural_t cpuCount;
	processor_info_array_t ia;
	mach_msg_type_number_t ic;

	kern_return_t error = host_processor_info(mach_host_self(),
		PROCESSOR_CPU_LOAD_INFO, &cpuCount, &ia, &ic);
	if (error) {
		return error;
	}

	processor_cpu_load_info_data_t* cpuLoadInfo =
		(processor_cpu_load_info_data_t*) ia;

	unsigned int all[4];
	// cpu_no,user,system,idle,nice
	for (int cpu=0; cpu<cpuCount; cpu++){
	  printf("%d,", cpu);
	  for (int i=0; i<4; i++){
		printf("%d", cpuLoadInfo[cpu].cpu_ticks[i]);
		all[i] = all[i] + cpuLoadInfo[cpu].cpu_ticks[i];
		if (i != 3){
		  printf(",");
		}else{
		  printf("\n");
		}
	  }
	}
	printf("99,");
	for (int i=0; i<4; i++){
	  printf("%d", all[i]);
	  if (i != 3){
		printf(",");
	  }else{
		printf("\n");
	  }
	}

	vm_deallocate(mach_task_self(), (vm_address_t)ia, ic);
	return error;
}
`
