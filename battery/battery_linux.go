// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

// TODO: ALLOW FOR USER TO CHOOSE THEIR OWN BATTERY
// TODO: CHECK IMPLEMENTATION OF chooseMainBattery, I only have one battery so I cannot test. It worked with only one battery.
// TODO: HAVE A LOT OF DIFFERENT DATA. TO GET ALL FILES AND THEIR VALUES ```grep . /sys/class/power_supply/BAT1/*```

package battery

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// Kernel info https://www.kernel.org/doc/Documentation/power/power_supply_class.txt
// BatteryStruct Ref: https://github.com/torvalds/linux/blob/24a19b953e2a3d0ea5a8c4951a0ab8f32720c907/include/linux/apm-emulation.h#L28
// Another Helpful resource: https://www.thelinuxvault.net/blog/how-to-obtain-a-laptop-battery-information-and-charging-state-on-linux/#3-sysclasspower_supply-kernel-level-battery-interface

func BatteryWithContext(ctx context.Context) (BatteryStruct, error) {
	var warns Warnings
	var BatteryData BatteryStruct

	var batteryAbsolutePath string

	// Get all the batteries
	allBatteries, err := getBatteries(ctx)
	if err != nil {
		warns.Add(err)
		return BatteryData, warns.Reference()
	}
	if len(allBatteries) == 1 {
		batteryAbsolutePath = allBatteries[0]
	} else {
		batteryAbsolutePath, err = chooseMainBattery(allBatteries, ctx)
		if err != nil {
			warns.Add(err)
		}
	}
	BatteryData.BatteryKey = filepath.Base(batteryAbsolutePath)
	batteryAbsolutePathDir := batteryAbsolutePath + "/*"

	batteryFiles, err := filepath.Glob(batteryAbsolutePathDir)
	if err != nil {
		return BatteryData, warns.Reference()
	}

	processFileDataInt := func(path string) (*int, error) {
		fileData, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fileDataStr := strings.TrimSuffix(string(fileData), "\n")
		fileDataInt, err := strconv.Atoi(fileDataStr)
		if err != nil {
			return nil, err
		}
		return &fileDataInt, nil
	}
	processFileDataString := func(path string) (*string, error) {
		fileData, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fileDataStr := strings.TrimSuffix(string(fileData), "\n")
		return &fileDataStr, nil
	}

	for _, file := range batteryFiles {
		fileName := filepath.Base(file)
		//var intData int

		// Match pre-declared files that we know what they do
		switch fileName {
		case "status":
			data, err := processFileDataString(batteryAbsolutePath + "/" + fileName)
			fmt.Println("Battery status:", data)
			BatteryData.Status = data
			if err != nil {
				warns.Add(err)
				continue
			}
			switch *data {
			case "Charging":
				result, err := calculateRemainingTime(batteryAbsolutePath, "charge")
				if err != nil {
					warns.Add(err)
					continue
				}
				BatteryData.RemainingTime = result
			case "Discharging":
				result, err := calculateRemainingTime(batteryAbsolutePath, "discharge")
				if err != nil {
					warns.Add(err)
					continue
				}
				BatteryData.RemainingTime = result
			default:
				negOne := float64(-1)
				BatteryData.RemainingTime = &negOne
			}
		case "technology":
			data, err := processFileDataString(batteryAbsolutePath + "/" + fileName)
			fmt.Println("Technology:", data)
			BatteryData.Technology = data
			if err != nil {
				warns.Add(err)
				continue
			}
		case "manufacturer":
			data, err := processFileDataString(batteryAbsolutePath + "/" + fileName)
			fmt.Println("Manufacturer:", data)
			BatteryData.Manufacturer = data
			if err != nil {
				warns.Add(err)
				continue
			}
		case "capacity_level":
			data, err := processFileDataString(batteryAbsolutePath + "/" + fileName)
			fmt.Println("BatteryLevelLang:", data)
			BatteryData.BatteryLevelLang = data
			if err != nil {
				warns.Add(err)
				continue
			}
		case "capacity":
			data, err := processFileDataInt(batteryAbsolutePath + "/" + fileName)
			fmt.Println("BatteryLevelInt:", data)
			BatteryData.BatteryLevelInt = data
			if err != nil {
				warns.Add(err)
				continue
			}

		case "current_now":
			data, err := processFileDataInt(batteryAbsolutePath + "/" + fileName)
			fmt.Println("current_now:", data)
			BatteryData.CurrentCharge = data
			if err != nil {
				warns.Add(err)
				continue
			}
		case "charge_control_start_threshold":
			data, err := processFileDataInt(batteryAbsolutePath + "/" + fileName)
			fmt.Println("BatteryMinimumPercentageLevelToStartCharging:", data)
			BatteryData.BatteryMinimumPercentageLevelToStartCharging = data
			if err != nil {
				warns.Add(err)
				continue
			}
		case "temp":
			strData, err := processFileDataString(batteryAbsolutePath + "/" + fileName)
			fmt.Println("temp:", strData)
			BatteryData.Temperature = strData
			if err != nil {
				warns.Add(err)
				continue
			}

			//case "energy_full_design":
			//	intData, err = processFileDataInt(batteryAbsolutePath + "/" + fileName)
			//	if err != nil {
			//		warns.Add(err)
			//		continue
			//	}
		}
	}

	return BatteryData, warns.Reference()
}

// Returns time in MS
// Returns 0 if fully charged
// Returns int (nullable) in milliseconds of time left, and possible error
//
// absolutePath = absolute path to battery, typeOfCalc is either "discharge" or "charge" depending on what you want to calculate. Discharge to time left till battery drain, charge to get time left till fully charged.
func calculateRemainingTime(absolutePath string, typeOfCalc string) (*float64, error) {
	processFileDataInt := func(path string) (int, error) {
		fileData, err := os.ReadFile(path)
		if err != nil {
			return 0, err
		}
		return strconv.Atoi(strings.TrimSpace(string(fileData)))
	}

	chargeFull, err := processFileDataInt(absolutePath + "/charge_full")
	if err != nil {
		return nil, err
	}

	chargeNow, err := processFileDataInt(absolutePath + "/charge_now")
	if err != nil {
		return nil, err
	}

	currentNow, err := processFileDataInt(absolutePath + "/current_now")
	if err != nil {
		return nil, err
	}

	if currentNow == 0 {
		return nil, errors.New("current_now is zero: unable to estimate charging time")
	}

	switch typeOfCalc {
	//Working
	case "charge":
		remaining := float64(chargeFull - chargeNow)
		if remaining <= 0 {
			zero := float64(0)
			return &zero, nil
		}
		hoursLeft := remaining / math.Abs(float64(currentNow))
		millisecondsLeft := hoursLeft * 3600000
		return &millisecondsLeft, nil

	//	Not accurate enough...
	case "discharge":
		voltageNow, err := processFileDataInt(absolutePath + "/voltage_now")
		if err != nil {
			return nil, err
		}

		// Convert to watts to be able to do math with remaining energy
		powerWatts := math.Abs(float64(voltageNow)) *
			math.Abs(float64(currentNow)) / 1e12

		if powerWatts == 0 {
			return nil, errors.New("power is zero: unable to estimate discharge time")
		}

		// calculate how much is left
		remainingEnergyWh := float64(chargeNow) *
			float64(voltageNow) / 1e12

		//how much is left compared to the draw
		hoursLeft := remainingEnergyWh / powerWatts
		millisecondsLeft := hoursLeft * 3600000

		return &millisecondsLeft, nil
	}

	return nil, errors.New("unable to calculate time")
}

// getBatteries
//
// []string: list of all BAT* dirs found in /sys/class/power_supply/*
//
// error: typically if error getting batteries dir or if there's nothing inside.
func getBatteries(ctx context.Context) ([]string, error) {
	var warn Warnings
	fmt.Println("Getting all batteries")

	var filepathContent []string

	// Get all content in the battery directory located at (absolute path) /sys/class/power_supply/
	batteriesAbsoluteFilePath := common.HostSysWithContext(ctx, "/class/power_supply/*")
	filepathContent, err := filepath.Glob(batteriesAbsoluteFilePath)
	if err != nil {
		return nil, err
	}
	// Return error if there's nothing inside the kernels battery dir
	if filepathContent == nil || len(filepathContent) == 0 {
		warn.Add(fmt.Errorf("no content found in %s", common.HostSysWithContext(ctx, "/class/power_supply/*")))
		return nil, err
	}

	// Get only the batteries with the kernel attribute of "Battery" inside it's `type` file
	var possibleBatteries []string
	for index := range filepathContent {
		item := filepathContent[index]
		if checkIfDirIsBattery(filepathContent[index]) {
			possibleBatteries = append(possibleBatteries, item)
		} else {
			continue
		}
	}

	// Return error if there's batteries found
	if possibleBatteries == nil || len(possibleBatteries) == 0 {
		return nil, fmt.Errorf("no batteries were found in %s", common.HostSysWithContext(ctx, "/class/power_supply/*"))
	}
	//If there's only one battery, perfect
	if len(possibleBatteries) == 1 {
		return possibleBatteries, nil
	}

	// Return a list of all Battery dirs
	return possibleBatteries, nil
}

// checkIfDirIsBattery reports whether absolutePath is a power supply
// directory for a battery, i.e. it contains a readable "type" file whose
// trimmed contents equal "Battery".
//
// The kernel exposes other power supply types ("UPS", "Mains", "USB") in the
// same location, so those return false. It also returns false if the file is
// missing or cannot be read.
//
// Example input: /sys/class/power_supply/BAT0
func checkIfDirIsBattery(absolutePath string) bool {
	fileData, err := os.ReadFile(absolutePath + "/type")
	if err != nil {
		return false
	}

	fileDataStr := strings.TrimSuffix(string(fileData), "\n")
	if string(fileDataStr) == "Battery" {
		return true
	}
	return false
}

// Some laptops have BAT1 as their primary internal battery, some have BAT0, just BAT, etc.
// In some laptops, multiple BAT batteries can appear, and this function chooses the battery with the higher amperage as the "main" battery and returns it.
// Return error if no batteries, or battery files cannot be read.
func chooseMainBattery(batteries []string, ctx context.Context) (string, error) {
	fmt.Println("Choosing Main Battery")

	var warn Warnings
	var largestBatteryMAHNumber int
	//0 default value
	var largestBattery string

	// FROM LINUX KERNEL https://www.kernel.org/doc/Documentation/power/power_supply_class.txt

	// Loop through all the batteries, and the one with the largest wattage is considered the "main" battery. Found in "energy_full_design" file
	//var largestBattery string
	for batteryIndex := range batteries {
		bytesToInt := func(bytes []byte) (int, error) {
			chargeFullDesignDataStr := string(bytes)
			chargeFullDesignDataStr = strings.TrimSuffix(string(bytes), "\n")
			return strconv.Atoi(chargeFullDesignDataStr)
		}

		checkIfNewLargestBattery := func(ampHours int) {
			if largestBatteryMAHNumber == 0 {
				largestBatteryMAHNumber = ampHours
				largestBattery = batteries[batteryIndex]

				//If largestBatteryMAHNumber is initialised, check if the current battery is larger than the initialised one
			} else if ampHours > largestBatteryMAHNumber {
				largestBatteryMAHNumber = ampHours
				largestBattery = batteries[batteryIndex]
			}
		}

		// Absolute path is something like /sys/class/power_supply/BAT1/energy_full_design
		// amp-hours
		chargeFullDesignData, err := os.ReadFile(batteries[batteryIndex] + "/charge_full_design")
		// If there's no error (meaning file is found)
		if err == nil {
			ampHours, err2 := bytesToInt(chargeFullDesignData)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				//If largestBatteryMAHNumber is not initialised
				checkIfNewLargestBattery(ampHours)
			}

		} else { //If file is not found or similar complications, go to secondary file "energy_full_design" + "voltage_now"

			// To calculate amp-hours, we need watt-hours and voltage. watt-hours is found in "energy_full_design", voltage is found in "voltage_now"
			// Absolute path is something like "/sys/class/power_supply/BAT1/energy_full_design"
			// watt-hours
			energyFullDesignData, err := os.ReadFile(batteries[batteryIndex] + "/energy_full_design")
			// If file cannot be found or opened
			//This therefore means that the battery has no information we can compare, so we must give up on this battery
			if err == nil {
				warn.Add(err)
				continue
			}

			// Absolute path is something like /sys/class/power_supply/BAT1/voltage_now
			// voltage_now
			voltageNow, err := os.ReadFile(batteries[batteryIndex] + "/voltage_now")
			// If file cannot be found or opened
			// If voltage isn't available, we can't compare, so we must give up on this battery
			if err == nil {
				warn.Add(err)
				continue
			}

			energyInt, err := bytesToInt(energyFullDesignData)
			if err != nil {
				fmt.Println(err)
				warn.Add(err)
				continue
			}
			voltageInt, err := bytesToInt(voltageNow)
			if err != nil {
				fmt.Println(err)
				warn.Add(err)
				continue
			}

			// Calculation: Amp Hours = Watt Hours ÷ Voltage
			ampHours := energyInt / voltageInt

			//Therefore, Check if the battery is the newest one
			checkIfNewLargestBattery(ampHours)

		}

	}

	return largestBattery, warn.Reference()
}
