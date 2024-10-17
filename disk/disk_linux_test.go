//go:build linux
// +build linux

package disk

import (
	"os"
	"testing"
)

func TestDiskStatsPost55(t *testing.T) {
	orig := os.Getenv("HOST_PROC")
	defer os.Setenv("HOST_PROC", orig)

	os.Setenv("HOST_PROC", "testdata/linux/diskstats-post5.5/proc")
	ios, err := IOCounters("sda", "sdc")
	if err != nil {
		t.Error("IOCounters failed")
	}

	expected := IOCountersStat{
		ReadCount:          256838102,
		MergedReadCount:    620512,
		ReadBytes:          5028599594 * sectorSize,
		ReadTime:           226563271,
		WriteCount:         418236058,
		MergedWriteCount:   7573415,
		WriteBytes:         8577305933 * sectorSize,
		WriteTime:          171833267,
		IopsInProgress:     0,
		IoTime:             141604084,
		WeightedIO:         402232601,
		DiscardCount:       168817,
		MergedDiscardCount: 110,
		DiscardBytes:       4991981424 * sectorSize,
		DiscardTime:        387582,
		FlushCount:         983197,
		FlushTime:          3448479,
		Name:               "sdc",
		SerialNumber:       "",
		Label:              "",
	}
	if ios["sdc"] != expected {
		t.Logf("IOCounterStats gave wrong results: expected %v actual %v", expected, ios["sdc"])
		t.Error("IOCounterStats gave wrong results")
	}
}

func TestDiskStatsPre55(t *testing.T) {
	orig := os.Getenv("HOST_PROC")
	os.Setenv("HOST_PROC", "testdata/linux/diskstats-pre5.5/proc")
	defer os.Setenv("HOST_PROC", orig)

	ios, err := IOCounters("sda", "sdc")
	if err != nil {
		t.Error("IOCounters failed")
	}
	expected := IOCountersStat{
		ReadCount:          256838102,
		MergedReadCount:    620512,
		ReadBytes:          5028599594 * sectorSize,
		ReadTime:           226563271,
		WriteCount:         418236058,
		MergedWriteCount:   7573415,
		WriteBytes:         8577305933 * sectorSize,
		WriteTime:          171833267,
		IopsInProgress:     0,
		IoTime:             141604084,
		WeightedIO:         402232601,
		DiscardCount:       168817,
		MergedDiscardCount: 110,
		DiscardBytes:       4991981424 * sectorSize,
		DiscardTime:        387582,
		FlushCount:         0,
		FlushTime:          0,
		Name:               "sdc",
		SerialNumber:       "",
		Label:              "",
	}
	if ios["sdc"] != expected {
		t.Logf("IOCounterStats gave wrong results: expected %v actual %v", expected, ios)
		t.Error("IOCounterStats gave wrong results")
	}

}

func TestDiskStatsPre418(t *testing.T) {
	orig := os.Getenv("HOST_PROC")
	defer os.Setenv("HOST_PROC", orig)

	os.Setenv("HOST_PROC", "testdata/linux/diskstats-pre4.18/proc")
	ios, err := IOCounters("sda", "sdc")
	if err != nil {
		t.Error("IOCounters failed")
	}
	expected := IOCountersStat{
		ReadCount:          256838102,
		MergedReadCount:    620512,
		ReadBytes:          5028599594 * sectorSize,
		ReadTime:           226563271,
		WriteCount:         418236058,
		MergedWriteCount:   7573415,
		WriteBytes:         8577305933 * sectorSize,
		WriteTime:          171833267,
		IopsInProgress:     0,
		IoTime:             141604084,
		WeightedIO:         402232601,
		DiscardCount:       0,
		MergedDiscardCount: 0,
		DiscardBytes:       0,
		DiscardTime:        0,
		FlushCount:         0,
		FlushTime:          0,
		Name:               "sdc",
		SerialNumber:       "",
		Label:              "",
	}
	if ios["sdc"] != expected {
		t.Logf("IOCounterStats gave wrong results: expected %v actual %v", expected, ios["sdc"])
		t.Error("IOCounterStats gave wrong results")
	}
}
