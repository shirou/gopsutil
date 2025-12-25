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
	ioKit, err := common.NewLibrary(common.IOKit)
	if err != nil {
		return nil, err
	}
	defer ioKit.Close()

	coreFoundation, err := common.NewLibrary(common.CoreFoundation)
	if err != nil {
		return nil, err
	}
	defer coreFoundation.Close()

	ta := &temperatureArm{
		ioKit:                     ioKit,
		cf:                        coreFoundation,
		cfRelease:                 common.GetFunc[common.CFReleaseFunc](coreFoundation, common.CFReleaseSym),
		cfStringCreateWithCString: common.GetFunc[common.CFStringCreateWithCStringFunc](coreFoundation, common.CFStringCreateWithCStringSym),
		cfArrayGetCount:           common.GetFunc[common.CFArrayGetCountFunc](coreFoundation, common.CFArrayGetCountSym),
		cfArrayGetValueAtIndex:    common.GetFunc[common.CFArrayGetValueAtIndexFunc](coreFoundation, common.CFArrayGetValueAtIndexSym),

		ioHIDEventSystemClientCreate:       common.GetFunc[common.IOHIDEventSystemClientCreateFunc](ioKit, common.IOHIDEventSystemClientCreateSym),
		ioHIDEventSystemClientSetMatching:  common.GetFunc[common.IOHIDEventSystemClientSetMatchingFunc](ioKit, common.IOHIDEventSystemClientSetMatchingSym),
		ioHIDEventSystemClientCopyServices: common.GetFunc[common.IOHIDEventSystemClientCopyServicesFunc](ioKit, common.IOHIDEventSystemClientCopyServicesSym),

		ioHIDServiceClientCopyProperty: common.GetFunc[common.IOHIDServiceClientCopyPropertyFunc](ioKit, common.IOHIDServiceClientCopyPropertySym),
		cfStringGetLength:              common.GetFunc[common.CFStringGetLengthFunc](coreFoundation, common.CFStringGetLengthSym),
		cfStringGetCString:             common.GetFunc[common.CFStringGetCStringFunc](coreFoundation, common.CFStringGetCStringSym),
		ioHIDServiceClientCopyEvent:    common.GetFunc[common.IOHIDServiceClientCopyEventFunc](ioKit, common.IOHIDServiceClientCopyEventSym),
		ioHIDEventGetFloatValue:        common.GetFunc[common.IOHIDEventGetFloatValueFunc](ioKit, common.IOHIDEventGetFloatValueSym),
		cfNumberCreate:                 common.GetFunc[common.CFNumberCreateFunc](coreFoundation, common.CFNumberCreateSym),
		cfDictionaryCreate:             common.GetFunc[common.CFDictionaryCreateFunc](coreFoundation, common.CFDictionaryCreateSym),
	}

	sensors := ta.matching(kHIDPageAppleVendor, kHIDPageAppleVendorTemperatureSensor)
	defer ta.cfRelease(uintptr(sensors))

	// Create HID system client
	system := ta.ioHIDEventSystemClientCreate(common.KCFAllocatorDefault)
	defer ta.cfRelease(uintptr(system))

	return ta.getSensors(system, sensors), nil
}

type temperatureArm struct {
	ioKit *common.Library
	cf    *common.Library

	cfRelease                 common.CFReleaseFunc
	cfStringCreateWithCString common.CFStringCreateWithCStringFunc
	cfArrayGetCount           common.CFArrayGetCountFunc
	cfArrayGetValueAtIndex    common.CFArrayGetValueAtIndexFunc
	cfStringGetLength         common.CFStringGetLengthFunc
	cfStringGetCString        common.CFStringGetCStringFunc
	cfNumberCreate            common.CFNumberCreateFunc
	cfDictionaryCreate        common.CFDictionaryCreateFunc

	ioHIDEventSystemClientCreate       common.IOHIDEventSystemClientCreateFunc
	ioHIDEventSystemClientSetMatching  common.IOHIDEventSystemClientSetMatchingFunc
	ioHIDEventSystemClientCopyServices common.IOHIDEventSystemClientCopyServicesFunc
	ioHIDServiceClientCopyProperty     common.IOHIDServiceClientCopyPropertyFunc
	ioHIDServiceClientCopyEvent        common.IOHIDServiceClientCopyEventFunc
	ioHIDEventGetFloatValue            common.IOHIDEventGetFloatValueFunc
}

func (ta *temperatureArm) getSensors(system, sensors unsafe.Pointer) []TemperatureStat {
	ta.ioHIDEventSystemClientSetMatching(uintptr(system), uintptr(sensors))
	matchingsrvs := ta.ioHIDEventSystemClientCopyServices(uintptr(system))

	if matchingsrvs == nil {
		return nil
	}
	defer ta.cfRelease(uintptr(matchingsrvs))

	str := ta.cfStr("Product")
	defer ta.cfRelease(uintptr(str))

	count := ta.cfArrayGetCount(uintptr(matchingsrvs))
	stats := make([]TemperatureStat, 0, count)

	nameSet := make(map[string]struct{})
	var i int32
	// traverse backward to keep the latest result first
	for i = count - 1; i >= 0; i-- {
		sc := ta.cfArrayGetValueAtIndex(uintptr(matchingsrvs), i)
		event := ta.ioHIDServiceClientCopyEvent(uintptr(sc), common.KIOHIDEventTypeTemperature, 0, 0)
		if event == nil {
			continue
		}

		temp := ta.ioHIDEventGetFloatValue(uintptr(event), ioHIDEventFieldBase(common.KIOHIDEventTypeTemperature))
		ta.cfRelease(uintptr(event))

		nameRef := ta.ioHIDServiceClientCopyProperty(uintptr(sc), uintptr(str))
		if nameRef != nil {
			buf := common.NewCStr(ta.cfStringGetLength(uintptr(nameRef)))
			ta.cfStringGetCString(uintptr(nameRef), buf, buf.Length(), common.KCFStringEncodingUTF8)

			name := buf.GoString()
			if _, ok := nameSet[name]; ok {
				ta.cfRelease(uintptr(nameRef))
				continue
			}

			stats = append(stats, TemperatureStat{
				SensorKey:   name,
				Temperature: temp,
			})
			nameSet[name] = struct{}{}
			ta.cfRelease(uintptr(nameRef))
		}
	}

	return stats
}

func (ta *temperatureArm) matching(page, usage int32) unsafe.Pointer {
	pageNum := ta.cfNumberCreate(common.KCFAllocatorDefault, common.KCFNumberIntType, uintptr(unsafe.Pointer(&page)))
	defer ta.cfRelease(uintptr(pageNum))

	usageNum := ta.cfNumberCreate(common.KCFAllocatorDefault, common.KCFNumberIntType, uintptr(unsafe.Pointer(&usage)))
	defer ta.cfRelease(uintptr(usageNum))

	k1 := ta.cfStr("PrimaryUsagePage")
	k2 := ta.cfStr("PrimaryUsage")

	defer ta.cfRelease(uintptr(k1))
	defer ta.cfRelease(uintptr(k2))

	keys := []unsafe.Pointer{k1, k2}
	values := []unsafe.Pointer{pageNum, usageNum}

	kCFTypeDictionaryKeyCallBacks, _ := ta.cf.Dlsym("kCFTypeDictionaryKeyCallBacks")
	kCFTypeDictionaryValueCallBacks, _ := ta.cf.Dlsym("kCFTypeDictionaryValueCallBacks")

	return ta.cfDictionaryCreate(common.KCFAllocatorDefault, &keys[0], &values[0], 2,
		kCFTypeDictionaryKeyCallBacks,
		kCFTypeDictionaryValueCallBacks)
}

func (ta *temperatureArm) cfStr(str string) unsafe.Pointer {
	return ta.cfStringCreateWithCString(common.KCFAllocatorDefault, str, common.KCFStringEncodingUTF8)
}

func ioHIDEventFieldBase(i int32) int32 {
	return i << 16
}
