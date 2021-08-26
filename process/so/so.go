// +build linux

package so

import (
	"regexp"
	"strconv"

	"github.com/DataDog/gopsutil/internal/common"
)

type libraryKey struct {
	Pathname       string // path of the library see by the process
	MountNameSpace ns     // namespace defined by the dev and inode
}

// Library define a dynamic library
type Library struct {
	libraryKey
	PidsPath []string // list of host pid path like /proc/<pid> per libraryKey
	HostPath string   // path of the library seen by the host
}

// AllLibraries represents a filter that matches all shared libraries
var AllLibraries = regexp.MustCompile(`\.so($|\.)`)

// Find returns the host-resolved paths of all shared libraries (per mount namespace) matching the given filter
// It does so by iterating over all /procPath/<PID>/maps and /procPath/<PID>/mountinfo files in the host
// If filter is nil, all entries from /procPath/<PID>/maps with a pathname are reported
func FindProc(procPath string, filter *regexp.Regexp) []Library {
	finder := newFinder(procPath)
	return finder.Find(filter)
}

// Find returns the host-resolved paths of all shared libraries (per mount namespace) matching the given filter
// It does so by iterating over all /proc/<PID>/maps and /proc/<PID>/mountinfo files in the host
// If filter is nil, all entries from /proc/<PID>/maps with a pathname are reported
func Find(filter *regexp.Regexp) []Library {
	return FindProc(common.HostProc(), filter)
}

// FromPID returns all shared libraries matching the given filter that are mapped into memory by a given PID
// If filter is nil, all entries from /proc/<PID>/maps with a pathname are reported
func FromPID(pid int32, filter *regexp.Regexp) []Library {
	return FindProc(common.HostProc(strconv.Itoa(int(pid))), filter)
}
