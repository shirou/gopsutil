// +build solaris

package cpu

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/mailru/easyjson"
	"github.com/shirou/gopsutil/internal/common"
)

// PIL_MAX has been 15 on Intel and Sparc since the launch of OpenSolaris:
// https://github.com/illumos/illumos-gate/blame/2428aad8462660fad2b105777063fea6f4192308/usr/src/uts/intel/sys/machlock.h#L107
// https://github.com/illumos/illumos-gate/blame/2428aad8462660fad2b105777063fea6f4192308/usr/src/uts/sparc/sys/machlock.h#L106
const PIL_MAX = 15

type TimesCPUStatInterruptLevel struct {
	Count uint64        // Number of times an interrupt at a given level was triggered
	Time  time.Duration // Time spent in interrupt context at this level
}

// TimesCPUStatInterrupt has Time and Counts for 15 Processor Interrupt Levels,
// zero-based indexing.
type TimesCPUStatInterrupt struct {
	CPUID uint
	PIL   []TimesCPUStatInterruptLevel // Processor Interrupt Level
}

func Times(perCPU bool) ([]TimesStat, error) {
	var ret []TimesStat

	inZone, err := common.InZone()
	if err != nil {
		return nil, err
	}

	switch {
	case perCPU:
		var cpuStats []TimesCPUStatSys
		cpuStats, err = TimesPerCPUSys()
		if err != nil {
			return nil, err
		}
		ret = make([]TimesStat, 0, len(cpuStats))

		for _, cpuStat := range cpuStats {
			ret = append(ret, TimesStat{
				CPU:       fmt.Sprintf("cpu%d", cpuStat.CPUID),
				User:      float64(time.Duration(cpuStat.CPUTimeUser) / time.Second),
				System:    float64(time.Duration(cpuStat.CPUTimeKernel) / time.Second),
				Idle:      float64(time.Duration(cpuStat.CPUTimeIdle) / time.Second),
				Nice:      0.0, // NOTE(seanc@): no value emitted by Illumos at present
				Iowait:    float64(time.Duration(cpuStat.IOWait) / time.Second),
				Irq:       float64(time.Duration(cpuStat.DeviceInterrupts) / time.Second),
				Softirq:   float64(time.Duration(cpuStat.CPUTimeIntr) / time.Second),
				Steal:     0.0, // NOTE(seanc@): no value emitted by Illumos at present
				Guest:     0.0, // NOTE(seanc@): no value emitted by Illumos at present
				GuestNice: 0.0, // NOTE(seanc@): no value emitted by Illumos at present
				Stolen:    0.0, // NOTE(seanc@): no value emitted by Illumos at present
			})
		}
	case !perCPU && inZone:
		var zoneStats []TimesZoneStat
		zoneStats, err = TimesZone()
		if err != nil {
			return nil, err
		}

		ret = make([]TimesStat, 0, len(zoneStats))

		for i, zone := range zoneStats {
			ret = append(ret, TimesStat{
				CPU:    fmt.Sprintf("zone%d", i),
				User:   float64(zone.UserTime / time.Second),
				System: float64(zone.SysTime / time.Second),
				Irq:    float64(zone.WaitIRQTime / time.Second),
			})
		}
	default:
		var cpuStats []TimesCPUStat
		cpuStats, err = TimesCPU()
		if err != nil {
			return nil, err
		}
		timeStat := TimesStat{}

		timeStat.CPU = "cpuN"
		for _, cpuStat := range cpuStats {
			timeStat.User += float64(time.Duration(cpuStat.CPUTimeUser) / time.Second)
			timeStat.System += float64(time.Duration(cpuStat.CPUTimeKernel) / time.Second)
			timeStat.Irq += float64(time.Duration(cpuStat.CPUTimeIntr) / time.Second)
		}

		ret = []TimesStat{timeStat}
	}

	return ret, nil
}

// TimesPerCPU returns the CPU time spent for each processor.  This information
// is extremely granular (and useful!).
func TimesPerCPU() ([]TimesCPUStat, error) {
	interruptStats, err := TimesPerCPUInterrupt()
	if err != nil {
		return nil, fmt.Errorf("cannot get CPU interrupt stats: %v", err)
	}

	sysStats, err := TimesPerCPUSys()
	if err != nil {
		return nil, fmt.Errorf("cannot get CPU system stats: %v", err)
	}

	vmStats, err := TimesPerCPUVM()
	if err != nil {
		return nil, fmt.Errorf("cannot get CPU VM stats: %v", err)
	}

	if len(interruptStats) != len(sysStats) && len(sysStats) != len(vmStats) {
		return nil, fmt.Errorf("length of stats not identical: invariant broken")
	}

	ret := make([]TimesCPUStat, len(interruptStats))
	for i := uint(0); i < uint(len(ret)); i++ {
		ret[i] = TimesCPUStat{
			CPUID: i,
			TimesCPUStatInterrupt: interruptStats[i],
			TimesCPUStatSys:       sysStats[i],
			TimesCPUStatVM:        vmStats[i],
		}
	}

	return ret, nil
}

func TimesPerCPUInterrupt() ([]TimesCPUStatInterrupt, error) {
	kstatPath, err := common.KStatPath()
	if err != nil {
		return nil, fmt.Errorf("cannot find kstat(1M): %v", err)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "cpu", "-n", "intrstat")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat: %v", err)
	}

	var kstatFrames _cpuInterruptKStatFrames
	if err := easyjson.Unmarshal(kstatJSON, &kstatFrames); err != nil {
		return nil, fmt.Errorf("cannot decode kstat(1M) CPU intrstat JSON: %v", err)
	}

	ret := make([]TimesCPUStatInterrupt, len(kstatFrames))
	for cpuNum, cpuPILStat := range kstatFrames {
		if cpuPILStat.Instance != uint(cpuNum) {
			return nil, fmt.Errorf("kstat(1M) CPU intrstat output not ordered by instance: invariant broken")
		}

		ret[cpuNum].CPUID = uint(cpuNum)

		pil := make([]TimesCPUStatInterruptLevel, PIL_MAX)
		psd := cpuPILStat.Data
		pil[0].Count = psd.Level1Count
		pil[0].Time = psd.Level1Time
		pil[1].Count = psd.Level2Count
		pil[1].Time = psd.Level2Time
		pil[2].Count = psd.Level3Count
		pil[2].Time = psd.Level3Time
		pil[3].Count = psd.Level4Count
		pil[3].Time = psd.Level4Time
		pil[4].Count = psd.Level5Count
		pil[4].Time = psd.Level5Time
		pil[5].Count = psd.Level6Count
		pil[5].Time = psd.Level6Time
		pil[6].Count = psd.Level7Count
		pil[6].Time = psd.Level7Time
		pil[7].Count = psd.Level8Count
		pil[7].Time = psd.Level8Time
		pil[8].Count = psd.Level9Count
		pil[8].Time = psd.Level9Time
		pil[9].Count = psd.Level10Count
		pil[9].Time = psd.Level10Time
		pil[10].Count = psd.Level11Count
		pil[10].Time = psd.Level11Time
		pil[11].Count = psd.Level12Count
		pil[11].Time = psd.Level12Time
		pil[12].Count = psd.Level13Count
		pil[12].Time = psd.Level13Time
		pil[13].Count = psd.Level14Count
		pil[13].Time = psd.Level14Time
		pil[14].Count = psd.Level15Count
		pil[14].Time = psd.Level15Time

		ret[cpuNum].PIL = pil
	}

	return ret, nil
}

func TimesPerCPUSys() ([]TimesCPUStatSys, error) {
	kstatPath, err := common.KStatPath()
	if err != nil {
		return nil, fmt.Errorf("cannot find kstat(1M): %v", err)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "cpu", "-n", "sys")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat: %v", err)
	}

	var kstatFrames _cpuSysKStatFrames
	if err := easyjson.Unmarshal(kstatJSON, &kstatFrames); err != nil {
		return nil, fmt.Errorf("cannot decode kstat(1M) zones JSON: %v", err)
	}

	ret := make([]TimesCPUStatSys, 0, len(kstatFrames))
	for cpuNum, cpuSysStat := range kstatFrames {
		if cpuSysStat.Instance != uint(cpuNum) {
			return nil, fmt.Errorf("kstat(1M) CPU sys output not ordered by instance: invariant broken")
		}

		cpuSysStat.Data.CPUID = uint(cpuNum)
		ret = append(ret, cpuSysStat.Data)
	}

	return ret, nil
}

func TimesPerCPUVM() ([]TimesCPUStatVM, error) {
	kstatPath, err := common.KStatPath()
	if err != nil {
		return nil, fmt.Errorf("cannot find kstat(1M): %v", err)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "cpu", "-n", "vm")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat: %v", err)
	}

	var kstatFrames _cpuVMKStatFrames
	if err := easyjson.Unmarshal(kstatJSON, &kstatFrames); err != nil {
		return nil, fmt.Errorf("cannot decode kstat(1M) CPU VM JSON: %v", err)
	}

	ret := make([]TimesCPUStatVM, 0, len(kstatFrames))
	for cpuNum, cpuVMStat := range kstatFrames {
		if cpuVMStat.Instance != uint(cpuNum) {
			return nil, fmt.Errorf("kstat(1M) CPU vm output not ordered by instance: invariant broken")
		}

		cpuVMStat.Data.CPUID = uint(cpuNum)
		ret = append(ret, cpuVMStat.Data)
	}

	return ret, nil
}

// TimesCPU returns the CPU time spent on all processors.  This information is
// extremely granular (and useful!).
func TimesCPU() ([]TimesCPUStat, error) {
	allTimes, err := TimesPerCPU()
	if err != nil {
		return nil, err
	}

	return allTimes, nil
}

// TimesZone returns the CPU time spent in the visible zone(s).  In a zone we
// can't get per-CPU accounting.  Instead, report back an accurate use of CPU
// time for the given zone(s) but only report one CPU, the "zoneN" CPU.
func TimesZone() ([]TimesZoneStat, error) {
	kstatPath, err := common.KStatPath()
	if err != nil {
		return nil, fmt.Errorf("cannot find kstat(1M): %v", err)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "zones")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat(1M): %v", err)
	}

	var kstatFrames _zonesKStatFrames
	if err := easyjson.Unmarshal(kstatJSON, &kstatFrames); err != nil {
		return nil, fmt.Errorf("cannot decode kstat(1M) zones JSON: %v", err)
	}

	ret := make([]TimesZoneStat, 0, len(kstatFrames))
	for _, zoneStat := range kstatFrames {
		zStat := zoneStat.Data.TimesZoneStat
		zStat.BootTime = time.Unix(zoneStat.Data.BootTime, 0)
		ret = append(ret, zStat)
	}

	return ret, nil
}

func Info() ([]InfoStat, error) {
	psrinfoPath, err := common.ProcessorInfoPath()
	if err != nil {
		return nil, fmt.Errorf("cannot find psrinfo(1M): %v", err)
	}

	psrinfoOut, err := invoke.Command(psrinfoPath, "-p", "-v")
	if err != nil {
		return nil, fmt.Errorf("cannot execute psrinfo(1M): %s", err)
	}

	procs, err := parseProcessorInfo(string(psrinfoOut))
	if err != nil {
		return nil, fmt.Errorf("error parsing psrinfo output: %s", err)
	}

	isainfoPath, err := common.ISAInfoPath()
	if err != nil {
		return nil, fmt.Errorf("cannot find isainfo(1): %v", err)
	}

	isainfoOut, err := invoke.Command(isainfoPath, "-b", "-v")
	if err != nil {
		return nil, fmt.Errorf("cannot execute isainfo(1): %s", err)
	}

	flags, err := parseISAInfo(string(isainfoOut))
	if err != nil {
		return nil, fmt.Errorf("error parsing isainfo(1) output: %s", err)
	}

	result := make([]InfoStat, 0, len(flags))
	for _, proc := range procs {
		procWithFlags := proc
		procWithFlags.Flags = flags
		result = append(result, procWithFlags)
	}

	return result, nil
}

var flagsMatch = regexp.MustCompile(`[\w\.]+`)

func parseISAInfo(cmdOutput string) ([]string, error) {
	words := flagsMatch.FindAllString(cmdOutput, -1)

	// Sanity check the output
	if len(words) < 4 || words[1] != "bit" || words[3] != "applications" {
		return nil, errors.New("attempted to parse invalid isainfo output")
	}

	flags := make([]string, len(words)-4)
	for i, val := range words[4:] {
		flags[i] = val
	}
	sort.Strings(flags)

	return flags, nil
}

var psrinfoMatch = regexp.MustCompile(`The physical processor has (?:([\d]+) virtual processor \(([\d]+)\)|([\d]+) cores and ([\d]+) virtual processors[^\n]+)\n(?:\s+ The core has.+\n)*\s+.+ \((\w+) ([\S]+) family (.+) model (.+) step (.+) clock (.+) MHz\)\n[\s]*(.*)`)

const (
	psrNumCoresOffset   = 1
	psrNumCoresHTOffset = 3
	psrNumHTOffset      = 4
	psrVendorIDOffset   = 5
	psrFamilyOffset     = 7
	psrModelOffset      = 8
	psrStepOffset       = 9
	psrClockOffset      = 10
	psrModelNameOffset  = 11
)

func parseProcessorInfo(cmdOutput string) ([]InfoStat, error) {
	matches := psrinfoMatch.FindAllStringSubmatch(cmdOutput, -1)

	var infoStatCount int32
	result := make([]InfoStat, 0, len(matches))
	for physicalIndex, physicalCPU := range matches {
		var step int32
		var clock float64

		if physicalCPU[psrStepOffset] != "" {
			stepParsed, err := strconv.ParseInt(physicalCPU[psrStepOffset], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for step as 32-bit integer: %s", physicalCPU[9], err)
			}
			step = int32(stepParsed)
		}

		if physicalCPU[psrClockOffset] != "" {
			clockParsed, err := strconv.ParseInt(physicalCPU[psrClockOffset], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for clock as 32-bit integer: %s", physicalCPU[10], err)
			}
			clock = float64(clockParsed)
		}

		var err error
		var numCores int64
		var numHT int64
		switch {
		case physicalCPU[psrNumCoresOffset] != "":
			numCores, err = strconv.ParseInt(physicalCPU[psrNumCoresOffset], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for core count as 32-bit integer: %s", physicalCPU[1], err)
			}

			for i := 0; i < int(numCores); i++ {
				result = append(result, InfoStat{
					CPU:        infoStatCount,
					PhysicalID: strconv.Itoa(physicalIndex),
					CoreID:     strconv.Itoa(i),
					Cores:      1,
					VendorID:   physicalCPU[psrVendorIDOffset],
					ModelName:  physicalCPU[psrModelNameOffset],
					Family:     physicalCPU[psrFamilyOffset],
					Model:      physicalCPU[psrModelOffset],
					Stepping:   step,
					Mhz:        clock,
				})
				infoStatCount++
			}
		case physicalCPU[psrNumCoresHTOffset] != "":
			numCores, err = strconv.ParseInt(physicalCPU[psrNumCoresHTOffset], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for core count as 32-bit integer: %s", physicalCPU[3], err)
			}

			numHT, err = strconv.ParseInt(physicalCPU[psrNumHTOffset], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for hyperthread count as 32-bit integer: %s", physicalCPU[4], err)
			}

			for i := 0; i < int(numCores); i++ {
				result = append(result, InfoStat{
					CPU:        infoStatCount,
					PhysicalID: strconv.Itoa(physicalIndex),
					CoreID:     strconv.Itoa(i),
					Cores:      int32(numHT) / int32(numCores),
					VendorID:   physicalCPU[psrVendorIDOffset],
					ModelName:  physicalCPU[psrModelNameOffset],
					Family:     physicalCPU[psrFamilyOffset],
					Model:      physicalCPU[psrModelOffset],
					Stepping:   step,
					Mhz:        clock,
				})
				infoStatCount++
			}
		default:
			return nil, errors.New("values for cores with and without hyperthreading are both set")
		}
	}
	return result, nil
}
