// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && arm64

package sensors

import (
	"context"
	"unsafe"

	"github.com/shirou/gopsutil/v4/internal/common"
)

const (
	kHIDPageAppleVendor                  = 0xff00
	kHIDPageAppleVendorTemperatureSensor = 5
)

func ReadTemperaturesArm() []TemperatureStat {
	temperatures, _ := TemperaturesWithContext(context.Background())
	return temperatures
}

func TemperaturesWithContext(_ context.Context) ([]TemperatureStat, error) {
	iokit, err := common.NewIOKitLib()
	if err != nil {
		return nil, err
	}
	defer iokit.Close()

	cf, err := common.NewCoreFoundationLib()
	if err != nil {
		return nil, err
	}
	defer cf.Close()

	ta := &temperatureArm{
		iokit: iokit,
		cf:    cf,
	}

	sensors := ta.matching(kHIDPageAppleVendor, kHIDPageAppleVendorTemperatureSensor)
	defer cf.CFRelease(uintptr(sensors))

	// Create HID system client
	system := iokit.IOHIDEventSystemClientCreate(common.KCFAllocatorDefault)
	defer cf.CFRelease(uintptr(system))

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
