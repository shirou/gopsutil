//
// gopsutil is a port of psutil(http://pythonhosted.org/psutil/).
// This covers these architectures.
//  - linux
//  - freebsd
//  - window
package gopsutil

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// ReadLines read contents from file and split by new line.
func ReadLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	line, err := r.ReadString('\n')
	for err == nil {
		ret = append(ret, strings.Trim(line, "\n"))
		line, err = r.ReadString('\n')
	}

	return ret, nil
}

func byteToString(orig []byte) string {
	n := -1
	l := -1
	for i, b := range orig {
		// skip left side null
		if l == -1 && b == 0 {
			continue
		}
		if l == -1 {
			l = i
		}

		if b == 0 {
			break
		}
		n = i + 1
	}
	if n == -1 {
		return string(orig)
	}
	return string(orig[l:n])
}

// Parse to int32 without error
func parseInt32(val string) int32 {
	vv, _ := strconv.ParseInt(val, 10, 32)
	return int32(vv)
}

// Parse to uint64 without error
func parseUint64(val string) uint64 {
	vv, _ := strconv.ParseInt(val, 10, 64)
	return uint64(vv)
}

// Check the target string slice containes src or not
func stringContains(target []string, src string) bool {
	for _, t := range target {
		if t == src {
			return true
		}
	}
	return false
}
