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
	"strings"
)

// Read contents from file and split by new line.
func ReadLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	ret := make([]string, 0)

	r := bufio.NewReader(f)
	line, err := r.ReadString('\n')
	for err == nil {
		ret = append(ret, strings.Trim(line, "\n"))
		line, err = r.ReadString('\n')
	}

	return ret, err
}

func byteToString(orig []byte) string {
	n := -1
	for i, b := range orig {
		if b == 0 {
			break
		}
		n = i + 1
	}
	if n == -1 {
		return string(orig)
	} else {
		return string(orig[:n])
	}

}
