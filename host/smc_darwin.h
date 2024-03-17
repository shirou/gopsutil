#ifndef __SMC_H__
#define __SMC_H__ 1

#include <IOKit/IOKitLib.h>

#define AMBIENT_AIR_0          "TA0P"
#define AMBIENT_AIR_1          "TA1P"
#define CPU_0_DIODE            "TC0D"
#define CPU_0_HEATSINK         "TC0H"
#define CPU_0_PROXIMITY        "TC0P"
#define ENCLOSURE_BASE_0       "TB0T"
#define ENCLOSURE_BASE_1       "TB1T"
#define ENCLOSURE_BASE_2       "TB2T"
#define ENCLOSURE_BASE_3       "TB3T"
#define GPU_0_DIODE            "TG0D"
#define GPU_0_HEATSINK         "TG0H"
#define GPU_0_PROXIMITY        "TG0P"
#define HARD_DRIVE_BAY         "TH0P"
#define MEMORY_SLOT_0          "TM0S"
#define MEMORY_SLOTS_PROXIMITY "TM0P"
#define NORTHBRIDGE            "TN0H"
#define NORTHBRIDGE_DIODE      "TN0D"
#define NORTHBRIDGE_PROXIMITY  "TN0P"
#define THUNDERBOLT_0          "TI0P"
#define THUNDERBOLT_1          "TI1P"
#define WIRELESS_MODULE        "TW0P"
#define CPU_EFFICIENCY_CORE_1  "Tp09"
#define CPU_EFFICIENCY_CORE_2  "Tp0T"
#define CPU_PERFORMANCE_CORE_1 "Tp01"
#define CPU_PERFORMANCE_CORE_2 "Tp05"
#define CPU_PERFORMANCE_CORE_3 "Tp0D"
#define CPU_PERFORMANCE_CORE_4 "Tp0H"
#define CPU_PERFORMANCE_CORE_5 "Tp0L"
#define CPU_PERFORMANCE_CORE_6 "Tp0P"
#define CPU_PERFORMANCE_CORE_7 "Tp0X"
#define CPU_PERFORMANCE_CORE_8 "Tp0b"
#define GPU_1                  "Tg05"
#define GPU_2                  "Tg0D"
#define GPU_3                  "Tg0L"
#define GPU_4                  "Tg0T"
#define AIRFLOW_LEFT           "TaLP"
#define AIRFLOW_RIGHT          "TaRF"
#define NAND                   "TH0x"
#define BATTERY_1              "TB1T"
#define BATTERY_2              "TB2T"
#define AIRPORT                "TW0P"

kern_return_t gopsutil_v3_open_smc(void);
kern_return_t gopsutil_v3_close_smc(void);
double gopsutil_v3_get_temperature(char *);

#endif // __SMC_H__
