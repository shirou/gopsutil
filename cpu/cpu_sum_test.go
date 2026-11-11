// SPDX-License-Identifier: BSD-3-Clause
package cpu

import "testing"

func TestSumTimesStatsTotalsPerCPUFields(t *testing.T) {
	per := []TimesStat{
		{CPU: "cpu0", User: 1, System: 2, Idle: 3, Nice: 4, Iowait: 5, Irq: 6, Softirq: 7, Steal: 8, Guest: 9, GuestNice: 10},
		{CPU: "cpu1", User: 10, System: 20, Idle: 30, Nice: 40, Iowait: 50, Irq: 60, Softirq: 70, Steal: 80, Guest: 90, GuestNice: 100},
	}
	got := sumTimesStats(per)
	if got.CPU != "cpu-total" {
		t.Fatalf("CPU name = %q, want cpu-total", got.CPU)
	}
	want := TimesStat{CPU: "cpu-total", User: 11, System: 22, Idle: 33, Nice: 44, Iowait: 55, Irq: 66, Softirq: 77, Steal: 88, Guest: 99, GuestNice: 110}
	if got != want {
		t.Fatalf("sumTimesStats() = %+v, want %+v", got, want)
	}
}

func TestSumTimesStatsEmpty(t *testing.T) {
	got := sumTimesStats(nil)
	want := TimesStat{CPU: "cpu-total"}
	if got != want {
		t.Fatalf("sumTimesStats(nil) = %+v, want %+v", got, want)
	}
}
