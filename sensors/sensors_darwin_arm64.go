// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && arm64

package sensors

import (
	"context"
	"unsafe"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func ReadTemperaturesArm() []TemperatureStat {
	temperatures, _ := TemperaturesWithContext(context.Background())
	return temperatures
}

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
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
		ioKit:                              ioKit,
		cf:                                 coreFoundation,
		cfRelease:                          common.GetFunc[common.CFReleaseFunc](coreFoundation, common.CFReleaseSym),
		cfStringCreateWithCString:          common.GetFunc[common.CFStringCreateWithCStringFunc](coreFoundation, common.CFStringCreateWithCStringSym),
		cfArrayGetCount:                    common.GetFunc[common.CFArrayGetCountFunc](coreFoundation, common.CFArrayGetCountSym),
		cfArrayGetValueAtIndex:             common.GetFunc[common.CFArrayGetValueAtIndexFunc](coreFoundation, common.CFArrayGetValueAtIndexSym),
		ioHIDEventSystemClientCreate:       common.GetFunc[common.IOHIDEventSystemClientCreateFunc](ioKit, common.IOHIDEventSystemClientCreateSym),
		ioHIDEventSystemClientSetMatching:  common.GetFunc[common.IOHIDEventSystemClientSetMatchingFunc](ioKit, common.IOHIDEventSystemClientSetMatchingSym),
		ioHIDEventSystemClientCopyServices: common.GetFunc[common.IOHIDEventSystemClientCopyServicesFunc](ioKit, common.IOHIDEventSystemClientCopyServicesSym),
	}

	ta.matching(0xff00, 5)
	defer ta.cfRelease(uintptr(ta.sensors))

	// Create HID system client
	system := ta.ioHIDEventSystemClientCreate(common.KCFAllocatorDefault)
	defer ta.cfRelease(uintptr(system))

	thermalNames := ta.getProductNames(system)
	thermalValues := ta.getThermalValues(system)
	result := dumpNameValues(thermalNames, thermalValues)

	return result, nil
}

func dumpNameValues(kvsN []string, kvsV []float64) []TemperatureStat {
	count := len(kvsN)
	temperatureMap := make(map[string]TemperatureStat)

	for i := 0; i < count; i++ {
		temperatureMap[kvsN[i]] = TemperatureStat{
			SensorKey:   kvsN[i],
			Temperature: kvsV[i],
		}
	}

	temperatures := make([]TemperatureStat, 0, len(temperatureMap))
	for _, stat := range temperatureMap {
		temperatures = append(temperatures, stat)
	}

	return temperatures
}

type temperatureArm struct {
	ioKit *common.Library
	cf    *common.Library

	cfRelease                 common.CFReleaseFunc
	cfStringCreateWithCString common.CFStringCreateWithCStringFunc
	cfArrayGetCount           common.CFArrayGetCountFunc
	cfArrayGetValueAtIndex    common.CFArrayGetValueAtIndexFunc

	ioHIDEventSystemClientCreate       common.IOHIDEventSystemClientCreateFunc
	ioHIDEventSystemClientSetMatching  common.IOHIDEventSystemClientSetMatchingFunc
	ioHIDEventSystemClientCopyServices common.IOHIDEventSystemClientCopyServicesFunc

	sensors unsafe.Pointer
}

func (ta *temperatureArm) getProductNames(system unsafe.Pointer) []string {
	ioHIDServiceClientCopyProperty := common.GetFunc[common.IOHIDServiceClientCopyPropertyFunc](ta.ioKit, common.IOHIDServiceClientCopyPropertySym)
	cfStringGetLength := common.GetFunc[common.CFStringGetLengthFunc](ta.cf, common.CFStringGetLengthSym)
	cfStringGetCString := common.GetFunc[common.CFStringGetCStringFunc](ta.cf, common.CFStringGetCStringSym)

	ta.ioHIDEventSystemClientSetMatching(uintptr(system), uintptr(ta.sensors))
	matchingsrvs := ta.ioHIDEventSystemClientCopyServices(uintptr(system))

	if matchingsrvs == nil {
		return nil
	}
	defer ta.cfRelease(uintptr(matchingsrvs))

	count := ta.cfArrayGetCount(uintptr(matchingsrvs))

	var i int32
	str := ta.cfStr("Product")
	defer ta.cfRelease(uintptr(str))

	names := make([]string, 0, count)
	for i = 0; i < count; i++ {
		sc := ta.cfArrayGetValueAtIndex(uintptr(matchingsrvs), i)
		name := ioHIDServiceClientCopyProperty(uintptr(sc), uintptr(str))

		if name != nil {
			buf := common.NewCStr(cfStringGetLength(uintptr(name)))
			cfStringGetCString(uintptr(name), buf, buf.Length(), common.KCFStringEncodingUTF8)

			names = append(names, buf.GoString())
			ta.cfRelease(uintptr(name))
		} else {
			// make sure the number of names and values are consistent
			names = append(names, "noname")
		}
	}

	return names
}

func (ta *temperatureArm) getThermalValues(system unsafe.Pointer) []float64 {
	ioHIDServiceClientCopyEvent := common.GetFunc[common.IOHIDServiceClientCopyEventFunc](ta.ioKit, common.IOHIDServiceClientCopyEventSym)
	ioHIDEventGetFloatValue := common.GetFunc[common.IOHIDEventGetFloatValueFunc](ta.ioKit, common.IOHIDEventGetFloatValueSym)

	ta.ioHIDEventSystemClientSetMatching(uintptr(system), uintptr(ta.sensors))
	matchingsrvs := ta.ioHIDEventSystemClientCopyServices(uintptr(system))

	if matchingsrvs == nil {
		return nil
	}
	defer ta.cfRelease(uintptr(matchingsrvs))

	count := ta.cfArrayGetCount(uintptr(matchingsrvs))

	var values []float64
	var i int32
	for i = 0; i < count; i++ {
		sc := ta.cfArrayGetValueAtIndex(uintptr(matchingsrvs), i)
		event := ioHIDServiceClientCopyEvent(uintptr(sc), common.KIOHIDEventTypeTemperature, 0, 0)
		temp := 0.0

		if event != nil {
			temp = ioHIDEventGetFloatValue(uintptr(event), ioHIDEventFieldBase(common.KIOHIDEventTypeTemperature))
			ta.cfRelease(uintptr(event))
		}

		values = append(values, temp)
	}

	return values
}

func (ta *temperatureArm) matching(page, usage int) {
	cfNumberCreate := common.GetFunc[common.CFNumberCreateFunc](ta.cf, common.CFNumberCreateSym)
	cfDictionaryCreate := common.GetFunc[common.CFDictionaryCreateFunc](ta.cf, common.CFDictionaryCreateSym)

	pageNum := cfNumberCreate(common.KCFAllocatorDefault, common.KCFNumberIntType, uintptr(unsafe.Pointer(&page)))
	defer ta.cfRelease(uintptr(pageNum))

	usageNum := cfNumberCreate(common.KCFAllocatorDefault, common.KCFNumberIntType, uintptr(unsafe.Pointer(&usage)))
	defer ta.cfRelease(uintptr(usageNum))

	k1 := ta.cfStr("PrimaryUsagePage")
	k2 := ta.cfStr("PrimaryUsage")

	defer ta.cfRelease(uintptr(k1))
	defer ta.cfRelease(uintptr(k2))

	keys := []unsafe.Pointer{k1, k2}
	values := []unsafe.Pointer{pageNum, usageNum}

	kCFTypeDictionaryKeyCallBacks, _ := ta.cf.Dlsym("kCFTypeDictionaryKeyCallBacks")
	kCFTypeDictionaryValueCallBacks, _ := ta.cf.Dlsym("kCFTypeDictionaryValueCallBacks")

	ta.sensors = cfDictionaryCreate(common.KCFAllocatorDefault, &keys[0], &values[0], 2,
		kCFTypeDictionaryKeyCallBacks,
		kCFTypeDictionaryValueCallBacks)
}

func (ta *temperatureArm) cfStr(str string) unsafe.Pointer {
	return ta.cfStringCreateWithCString(common.KCFAllocatorDefault, str, common.KCFStringEncodingUTF8)
}

func ioHIDEventFieldBase(i int32) int32 {
	return i << 16
}
