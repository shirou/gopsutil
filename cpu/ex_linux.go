// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package cpu

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// TimesStatEx contains the unscaled Linux CPU time counters from /proc/stat.
// Counters are in USER_HZ ticks, rather than the seconds used by TimesStat.
// They are reported as read: in particular, Iowait can decrease on Linux.
type TimesStatEx struct {
	CPU       string `json:"cpu"`
	User      uint64 `json:"user"`
	System    uint64 `json:"system"`
	Idle      uint64 `json:"idle"`
	Nice      uint64 `json:"nice"`
	Iowait    uint64 `json:"iowait"`
	Irq       uint64 `json:"irq"`
	Softirq   uint64 `json:"softirq"`
	Steal     uint64 `json:"steal"`
	Guest     uint64 `json:"guest"`
	GuestNice uint64 `json:"guestNice"`
}

func (c TimesStatEx) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

type ExLinux struct{}

func NewExLinux() *ExLinux {
	return &ExLinux{}
}

func (ex *ExLinux) Times(percpu bool) ([]TimesStatEx, error) {
	return ex.TimesWithContext(context.Background(), percpu)
}

// TimesWithContext returns raw CPU tick counters without floating-point conversion.
// CPU names and selection match the package-level Times function; absent optional counters are zero.
// Unlike the legacy API, read errors and invalid selected CPU rows are returned.
func (*ExLinux) TimesWithContext(ctx context.Context, percpu bool) ([]TimesStatEx, error) {
	file, err := os.Open(common.HostProcWithContext(ctx, "stat"))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	ret := []TimesStatEx{}
	if percpu {
		if _, err := reader.ReadString('\n'); err != nil {
			if errors.Is(err, io.EOF) {
				return ret, nil
			}
			return nil, err
		}
	}
	for {
		if percpu {
			prefix, err := reader.Peek(3)
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}
			if string(prefix) != "cpu" {
				return ret, nil
			}
		}
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		if line == "" {
			return ret, nil
		}
		stat, parseErr := parseStatLineEx(line)
		if parseErr != nil {
			return nil, parseErr
		}
		ret = append(ret, stat)
		if !percpu || errors.Is(err, io.EOF) {
			return ret, nil
		}
	}
}

func parseStatLineEx(line string) (TimesStatEx, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return TimesStatEx{}, errors.New("stat does not contain cpu info")
	}
	if !strings.HasPrefix(fields[0], "cpu") {
		return TimesStatEx{}, errors.New("not contain cpu")
	}
	stat := TimesStatEx{CPU: fields[0]}
	if stat.CPU == "cpu" {
		stat.CPU = "cpu-total"
	}
	counters := []*uint64{
		&stat.User, &stat.Nice, &stat.System, &stat.Idle, &stat.Iowait,
		&stat.Irq, &stat.Softirq, &stat.Steal, &stat.Guest, &stat.GuestNice,
	}
	for i, counter := range counters {
		if i+1 >= len(fields) {
			break
		}
		value, err := strconv.ParseUint(fields[i+1], 10, 64)
		if err != nil {
			return TimesStatEx{}, fmt.Errorf("parse CPU time field %d for %s: %w", i+1, stat.CPU, err)
		}
		*counter = value
	}
	return stat, nil
}
