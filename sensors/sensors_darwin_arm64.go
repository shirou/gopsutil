// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && arm64

package sensors

import (
	"context"
	"sync"
	"unsafe"

	"github.com/shirou/gopsutil/v4/internal/common"
)

const (
	kHIDPageAppleVendor                  = 0xff00
	kHIDPageAppleVendorTemperatureSensor = 5
)

// Keep IOKit and CoreFoundation libraries open for the process lifetime.
// Opening and closing them on every call causes SIGBUS/SIGSEGV crashes
// because the Go runtime (GC, timers) can interact with invalidated
// library handles after Dlclose.
// See: https://github.com/shirou/gopsutil/issues/1832
var (
	sensorLibOnce sync.Once
	sensorIOKit   *common.IOKitLib
	sensorCF      *common.CoreFoundationLib
	sensorLibErr  error
)

func initSensorLibraries() {
	sensorIOKit, sensorLibErr = common.NewIOKitLib()
	if sensorLibErr != nil {
		return
	}
	sensorCF, sensorLibErr = common.NewCoreFoundationLib()
}

func ReadTemperaturesArm() []TemperatureStat {
	temperatures, _ := TemperaturesWithContext(context.Background())
	return temperatures
}

func TemperaturesWithContext(_ context.Context) ([]TemperatureStat, error) {
	sensorLibOnce.Do(initSensorLibraries)
	if sensorLibErr != nil {
		return nil, sensorLibErr
	}

	ta := &temperatureArm{
		iokit: sensorIOKit,
		cf:    sensorCF,
	}

	sensors := ta.matching(kHIDPageAppleVendor, kHIDPageAppleVendorTemperatureSensor)
	defer sensorCF.CFRelease(uintptr(sensors))

	// Create HID system client
	system := sensorIOKit.IOHIDEventSystemClientCreate(common.KCFAllocatorDefault)
	defer sensorCF.CFRelease(uintptr(system))

	return ta.getSensors(system, sensors), nil
}

type temperatureArm struct {
	iokit *common.IOKitLib
	cf    *common.CoreFoundationLib
}

func (ta *temperatureArm) getSensors(system, sensors unsafe.Pointer) []TemperatureStat {
	ta.iokit.IOHIDEventSystemClientSetMatching(uintptr(system), uintptr(sensors))
	matchingsrvs := ta.iokit.IOHIDEventSystemClientCopyServices(uintptr(system))

	if matchingsrvs == nil {
		return nil
	}
	defer ta.cf.CFRelease(uintptr(matchingsrvs))

	str := ta.cf.CFStringCreateWithCString(common.KCFAllocatorDefault, "Product", common.KCFStringEncodingUTF8)
	defer ta.cf.CFRelease(uintptr(str))

	count := ta.cf.CFArrayGetCount(uintptr(matchingsrvs))
	stats := make([]TemperatureStat, 0, count)

	nameSet := make(map[string]struct{})
	for i := count - 1; i >= 0; i-- {
		sc := ta.cf.CFArrayGetValueAtIndex(uintptr(matchingsrvs), i)
		event := ta.iokit.IOHIDServiceClientCopyEvent(uintptr(sc), common.KIOHIDEventTypeTemperature, 0, 0)
		if event == nil {
			continue
		}

		temp := ta.iokit.IOHIDEventGetFloatValue(uintptr(event), ioHIDEventFieldBase(common.KIOHIDEventTypeTemperature))
		ta.cf.CFRelease(uintptr(event))

		nameRef := ta.iokit.IOHIDServiceClientCopyProperty(uintptr(sc), uintptr(str))
		if nameRef != nil {
			buf := common.NewCStr(common.GetCFStringBufLengthForUTF8(ta.cf.CFStringGetLength(uintptr(nameRef))))
			ta.cf.CFStringGetCString(uintptr(nameRef), buf, buf.Length(), common.KCFStringEncodingUTF8)

			name := buf.GoString()
			if _, ok := nameSet[name]; ok {
				ta.cf.CFRelease(uintptr(nameRef))
				continue
			}

			stats = append(stats, TemperatureStat{
				SensorKey:   name,
				Temperature: temp,
			})
			nameSet[name] = struct{}{}
			ta.cf.CFRelease(uintptr(nameRef))
		}
	}

	return stats
}

func (ta *temperatureArm) matching(page, usage int32) unsafe.Pointer {
	pageNum := ta.cf.CFNumberCreate(common.KCFAllocatorDefault, common.KCFNumberIntType, uintptr(unsafe.Pointer(&page)))
	defer ta.cf.CFRelease(uintptr(pageNum))

	usageNum := ta.cf.CFNumberCreate(common.KCFAllocatorDefault, common.KCFNumberIntType, uintptr(unsafe.Pointer(&usage)))
	defer ta.cf.CFRelease(uintptr(usageNum))

	k1 := ta.cf.CFStringCreateWithCString(common.KCFAllocatorDefault, "PrimaryUsagePage", common.KCFStringEncodingUTF8)
	k2 := ta.cf.CFStringCreateWithCString(common.KCFAllocatorDefault, "PrimaryUsage", common.KCFStringEncodingUTF8)

	defer ta.cf.CFRelease(uintptr(k1))
	defer ta.cf.CFRelease(uintptr(k2))

	keys := []unsafe.Pointer{k1, k2}
	values := []unsafe.Pointer{pageNum, usageNum}

	kCFTypeDictionaryKeyCallBacks, _ := ta.cf.Dlsym("kCFTypeDictionaryKeyCallBacks")
	kCFTypeDictionaryValueCallBacks, _ := ta.cf.Dlsym("kCFTypeDictionaryValueCallBacks")

	return ta.cf.CFDictionaryCreate(common.KCFAllocatorDefault, &keys[0], &values[0], 2,
		kCFTypeDictionaryKeyCallBacks,
		kCFTypeDictionaryValueCallBacks)
}

func ioHIDEventFieldBase(i int32) int32 {
	return i << 16
}
