// +build linux

package so

import (
	"bufio"
	"os"
	"path"
	fp "path/filepath"
	"regexp"
	"strconv"

	"github.com/DataDog/gopsutil/process/filepath"
)

type finder struct {
	procRoot     string
	pathResolver *filepath.Resolver
	buffer       *bufio.Reader
}

func newFinder(procRoot string) *finder {
	buffer := bufio.NewReader(nil)
	realProcRoot := procRoot
	/* if /proc/<pid> is passed directly, use the dirname() as proc root */
	_, err := strconv.Atoi(path.Base(procRoot))
	if err == nil {
		realProcRoot = path.Dir(procRoot)
	}
	return &finder{
		procRoot:     procRoot,
		pathResolver: filepath.NewResolver(realProcRoot),
		buffer:       buffer,
	}
}

func (f *finder) Find(filter *regexp.Regexp) (result []Library) {
	mapLib := make(map[libraryKey]Library)
	err := iteratePIDS(f.procRoot, func(pidPath string, info os.FileInfo, mntNS ns) {
		libs := getSharedLibraries(pidPath, f.buffer, filter)
		f.pathResolver.LoadPIDMounts(pidPath)

		for _, lib := range libs {
			k := libraryKey{
				Pathname:       lib,
				MountNameSpace: mntNS,
			}
			if m, ok := mapLib[k]; ok {
				m.PidsPath = append(m.PidsPath, pidPath)
				mapLib[k] = m
				continue
			}

			hostPath := f.pathResolver.Resolve(lib)
			if hostPath == "" {
				continue
			}

			mapLib[k] = Library{
				libraryKey: k,
				HostPath:   hostPath,
				PidsPath:   []string{pidPath},
			}
		}
	})
	if err != nil {
		return result
	}
	for _, l := range mapLib {
		result = append(result, l)
	}
	return result
}

func iteratePIDS(procRoot string, fn callback) error {
	w := newWalker(procRoot, fn)
	return fp.Walk(procRoot, fp.WalkFunc(w.walk))
}
