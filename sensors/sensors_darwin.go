// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && !arm64

package sensors

import (
	"context"
	"errors"
	"unsafe"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	ioKit, err := common.NewLibrary(common.IOKit)
	if err != nil {
		return nil, err
	}
	defer ioKit.Close()

	smc, err := common.NewSMC(ioKit)
	if err != nil {
		return nil, err
	}
	defer smc.Close()

	temperatures := make([]TemperatureStat, 0, len(temperatureKeys))
	for _, key := range temperatureKeys {
		temperatures = append(temperatures, TemperatureStat{
			SensorKey:   key,
			Temperature: getTemperature(smc, key),
		})
	}

	return temperatures, nil
}

var temperatureKeys = []string{
	"TA0P", // AMBIENT_AIR_0
	"TA1P", // AMBIENT_AIR_1
	"TC0D", // CPU_0_DIODE
	"TC0H", // CPU_0_HEATSINK
	"TC0P", // CPU_0_PROXIMITY
	"TB0T", // ENCLOSURE_BASE_0
	"TB1T", // ENCLOSURE_BASE_1
	"TB2T", // ENCLOSURE_BASE_2
	"TB3T", // ENCLOSURE_BASE_3
	"TG0D", // GPU_0_DIODE
	"TG0H", // GPU_0_HEATSINK
	"TG0P", // GPU_0_PROXIMITY
	"TH0P", // HARD_DRIVE_BAY
	"TM0S", // MEMORY_SLOT_0
	"TM0P", // MEMORY_SLOTS_PROXIMITY
	"TN0H", // NORTHBRIDGE
	"TN0D", // NORTHBRIDGE_DIODE
	"TN0P", // NORTHBRIDGE_PROXIMITY
	"TI0P", // THUNDERBOLT_0
	"TI1P", // THUNDERBOLT_1
	"TW0P", // WIRELESS_MODULE
}

type smcReturn struct {
	data     [32]uint8
	dataType uint32
	dataSize uint32
	kSMC     uint8
}

type smcPLimitData struct {
	version   uint16
	length    uint16
	cpuPLimit uint32
	gpuPLimit uint32
	memPLimit uint32
}

type smcKeyInfoData struct {
	dataSize       uint32
	dataType       uint32
	dataAttributes uint8
}

type smcVersion struct {
	major    byte
	minor    byte
	build    byte
	reserved byte
	release  uint16
}

type smcParamStruct struct {
	key        uint32
	vers       smcVersion
	plimitData smcPLimitData
	keyInfo    smcKeyInfoData
	result     uint8
	status     uint8
	data8      uint8
	data32     uint32
	bytes      [32]byte
}

const (
	smcKeySize   = 4
	dataTypeSp78 = "sp78"
)

func getTemperature(smc *common.SMC, key string) float64 {
	result, err := readSMC(smc, key)
	if err != nil {
		return 0.0
	}

	if result.dataSize == 2 && result.dataType == toUint32(dataTypeSp78) {
		return 0.0
	}

	return float64(result.data[0])
}

func readSMC(smc *common.SMC, key string) (*smcReturn, error) {
	input := new(smcParamStruct)
	resultSmc := new(smcReturn)

	input.key = toUint32(key)
	input.data8 = common.KSMCGetKeyInfo

	result, err := callSMC(smc, input)
	resultSmc.kSMC = result.result

	if err != nil || result.result != common.KSMCSuccess {
		return resultSmc, errors.New("ERROR: IOConnectCallStructMethod failed")
	}

	resultSmc.dataSize = uint32(result.keyInfo.dataSize)
	resultSmc.dataType = uint32(result.keyInfo.dataSize)

	input.keyInfo.dataSize = result.keyInfo.dataSize
	input.data8 = common.KSMCReadKey

	result, err = callSMC(smc, input)
	resultSmc.kSMC = result.result

	if err != nil || result.result != common.KSMCSuccess {
		return resultSmc, err
	}

	resultSmc.data = result.bytes
	return resultSmc, nil
}

func callSMC(smc *common.SMC, input *smcParamStruct) (*smcParamStruct, error) {
	output := new(smcParamStruct)
	inputCnt := unsafe.Sizeof(*input)
	outputCnt := unsafe.Sizeof(*output)

	result := smc.CallStruct(common.KSMCHandleYPCEvent,
		uintptr(unsafe.Pointer(input)), inputCnt, uintptr(unsafe.Pointer(output)), &outputCnt)

	if result != 0 {
		return output, errors.New("ERROR: IOConnectCallStructMethod failed")
	}

	return output, nil
}

func toUint32(key string) uint32 {
	if len(key) != smcKeySize {
		return 0
	}

	var ans uint32
	var shift uint32 = 24

	for i := 0; i < smcKeySize; i++ {
		ans += uint32(key[i]) << shift
		shift -= 8
	}

	return ans
}
