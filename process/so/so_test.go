// +build linux

package so

import (
	"strconv"
	"testing"

	"github.com/DataDog/gopsutil/process/so/testutil"
	"github.com/stretchr/testify/assert"
)

func TestForeachPIDNoFilter(t *testing.T) {
	for _, p := range testutil.ProcFS {
		t.Log("testing PID :", p.Pid)

		libs := FindProc("testdata/procfs/"+strconv.Itoa(p.Pid), nil)
		libPathname := []string{}
		for _, lib := range libs {
			libPathname = append(libPathname, lib.Pathname)
		}

		expected := p.Libraries
		assert.ElementsMatch(t, expected, libPathname)
	}
}

func TestForeachPIDAllLibrariesFilter(t *testing.T) {
	for _, p := range testutil.ProcFS {
		t.Log("testing PID :", p.Pid)

		libs := FindProc("testdata/procfs/"+strconv.Itoa(p.Pid), AllLibraries)
		libPathname := []string{}
		for _, lib := range libs {
			libPathname = append(libPathname, lib.Pathname)
		}

		expected := []string{}
		for _, lib := range p.Libraries {
			if AllLibraries.MatchString(lib) {
				expected = append(expected, lib)
			}
		}
		assert.ElementsMatch(t, expected, libPathname)
	}
}
