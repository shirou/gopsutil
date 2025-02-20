// SPDX-License-Identifier: BSD-3-Clause
//go:build solaris

package sensors

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
)

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	var ret []TemperatureStat

	out, err := invoke.CommandWithContext(ctx, "ipmitool", "-c", "sdr", "list")
	if err != nil {
		return ret, err
	}

	r := csv.NewReader(strings.NewReader(string(out)))
	// Output may contain errors, e.g. "bmc_send_cmd: Permission denied", don't expect a consistent number of records
	r.FieldsPerRecord = -1
	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return ret, err
		}
		// CPU1 Temp,40,degrees C,ok
		if len(record) < 3 || record[1] == "" || record[2] != "degrees C" {
			continue
		}
		v, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return ret, err
		}
		ts := TemperatureStat{
			SensorKey:   strings.TrimSuffix(record[0], " Temp"),
			Temperature: v,
		}
		ret = append(ret, ts)
	}

	return ret, nil
}
