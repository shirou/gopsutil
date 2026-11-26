// MAIN SOURCE: https://kernel.googlesource.com/pub/scm/linux/kernel/git/torvalds/linux/+/refs/tags/v7.3-rc7/Documentation/ABI/testing/sysfs-class-power
// https://github.com/torvalds/linux/blob/24a19b953e2a3d0ea5a8c4951a0ab8f32720c907/include/linux/apm-emulation.h#L28

package battery

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

type Warnings = common.Warnings

var invoke common.Invoker = common.Invoke{}

// BatteryStruct
// Each attribute is a pointer because it may not exist, this is dependent on system.
//
// Not sure if this is the best way to do this, only real way I found of making a go struct nullable?
type BatteryStruct struct {
	// BatteryKey is the battery identifier/name.
	//
	// # Linux
	//
	// The battery's directory name in /sys/class/power_supply/ (for example, BAT0).
	BatteryKey string `json:"batteryKey"`

	// Status is the current state of the battery.
	//
	// # Linux
	//
	// One of: Charging, Discharging, Full, Not charging.
	//
	// Found in:
	//
	//	/sys/class/power_supply/<BatteryKey>/status
	Status *string `json:"status,omitempty"`

	// Technology is the battery chemistry, for example "Li-ion".
	//
	// # Linux
	//
	// Found in:
	//
	//	/sys/class/power_supply/<BatteryKey>/technology
	Technology *string `json:"technology,omitempty"`

	// Manufacturer is the battery manufacturer name.
	//
	// # Linux
	//
	// Found in:
	//
	//	/sys/class/power_supply/<BatteryKey>/manufacturer
	Manufacturer *string `json:"manufacturer,omitempty"`

	// BatteryLevelLang is the battery level described in words.
	//
	// # Linux
	//
	// One of: Unknown, Critical, Low, Normal, High, Full.
	//
	// Found in:
	//
	//	/sys/class/power_supply/<BatteryKey>/capacity_level
	BatteryLevelLang *string `json:"batteryLevelLang,omitempty"`

	// BatteryLevelInt is the current battery level as a percentage (0-100).
	//
	// # Linux
	//
	// Found in:
	//
	//	/sys/class/power_supply/<BatteryKey>/capacity
	BatteryLevelInt *int `json:"batteryLevelInt,omitempty"`

	// CurrentCharge
	//
	// # Linux
	//
	// "Reports an instant, single IBAT current reading for the battery. This value is not averaged/smoothed"
	// https://github.com/torvalds/linux/blob/24a19b953e2a3d0ea5a8c4951a0ab8f32720c907/Documentation/ABI/testing/sysfs-class-power#L41
	//
	//	/sys/class/power_supply/<BatteryKey>/current_now
	CurrentCharge *int `json:"currentCharge,omitempty"`

	// BatteryMinimumPercentageLevelToStartCharging is the charge percentage
	// (0-100) below which the battery starts charging.
	//
	// # Linux
	//
	// Found in:
	//
	//	/sys/class/power_supply/<BatteryKey>/charge_control_start_threshold
	BatteryMinimumPercentageLevelToStartCharging *int `json:"batteryMinimumPercentageLevelToStartCharging,omitempty"`

	// Temperature is the reported temperature of the battery.
	//
	// # Linux
	//
	// Measured in 1/10ths of a degree Celsius.
	//
	// Found in:
	//
	//	/sys/class/power_supply/<BatteryKey>/temp
	Temperature *string `json:"temperature,omitempty"`

	// RemainingTime is the time remaining to either discharge or charge in Milliseconds. -1 Seconds for "full charge" or "not charging" state
	//
	// # Linux
	//
	// Calculated with (3600000 -> milliseconds in an hour):
	//
	// Charge Time := (PowerSupplyChargeFull - PowerSupplyChargeNow) / PowerSupplyCurrentNow * 3600000
	//
	// Discharge Time := PowerSupplyChargeNow/ PowerSupplyCurrentNow * 3600000
	RemainingTime *float64 `json:"remainingTimeCharge,omitempty"`
}

func BatteryInfo() (BatteryStruct, error) {
	batteryData, err := BatteryWithContext(context.Background())
	return batteryData, err
}
