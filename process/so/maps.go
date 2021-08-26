// +build linux

package so

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"

	"github.com/DataDog/gopsutil/internal/common"
)

const soPathnameColumnIdx = 5

func getSharedLibraries(pidPath string, b *bufio.Reader, filter *regexp.Regexp) []string {
	f, err := os.Open(filepath.Join(pidPath, "/maps"))
	if err != nil {
		return nil
	}
	defer f.Close()
	b.Reset(f)

	return parseMaps(b, filter)
}

// parseMaps takes in an bufio.Reader representing a memory mapping
// file from the procfs (eg. /proc/<PID>/maps) and extracts the shared library names from it
// that match the given filter. If filter is nil all entries are returned.
//
// Example:
// 7f135146b000-7f135147a000 r--p 00000000 fd:00 268743 /usr/lib/x86_64-linux-gnu/libm-2.31.so
// 7f135147a000-7f1351521000 r-xp 0000f000 fd:00 268743 /usr/lib/x86_64-linux-gnu/libm-2.31.so
// 7f1351521000-7f13515b8000 r--p 000b6000 fd:00 268743 /usr/lib/x86_64-linux-gnu/libm-2.31.so
// 7f13515b8000-7f13515b9000 r--p 0014c000 fd:00 268743 /usr/lib/x86_64-linux-gnu/libm-2.31.so
//
// Would return ["/usr/lib/x86_64-linux-gnu/libm-2.31.so"]
func parseMaps(r *bufio.Reader, filter *regexp.Regexp) (libs []string) {
	set := common.NewStringSet()
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}

		field := bytes.Fields(line)
		if len(field) < 6 {
			continue
		}

		entry := field[soPathnameColumnIdx]
		set.Add(string(entry))
	}

	for _, lib := range set.GetAll() {
		if filter == nil || filter.MatchString(lib) {
			libs = append(libs, lib)
		}
	}

	return libs
}
