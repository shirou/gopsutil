// +build linux

package filepath

import (
	"bufio"
	fp "path/filepath"
	"strings"
)

// Resolver is responsible for translating namespaced file paths (eg.
// from a container point of view) to the root-namespace point of view
type Resolver struct {
	buffer *bufio.Reader

	root     *mountInfo
	nsMounts *mountInfo
}

// NewResolver returns a new `Resolver`
func NewResolver(procRoot string) *Resolver {
	initPIDPath := fp.Join(procRoot, "1")
	b := bufio.NewReader(nil)
	return &Resolver{
		buffer: b,
		root:   getMountInfo(initPIDPath, b),
	}
}

// Resolve a path from a potentially namespaced process to the host path
// Note that `LoadPIDMounts` must be called before this method
func (p *Resolver) Resolve(path string) string {
	if p.root == nil || p.nsMounts == nil || path == "" {
		return ""
	}

	nsMount := p.nsMounts.GetMount(path)
	if nsMount == nil {
		return ""
	}
	nsRelPath, err := fp.Rel(nsMount.mountPoint, path)
	if err != nil {
		return ""
	}

	var parentMount *mount
	for _, rootMount := range p.root.mounts {
		if rootMount.dev == nsMount.dev && strings.HasPrefix(nsMount.root, rootMount.root) {
			parentMount = rootMount
			break
		}
	}

	if parentMount == nil {
		return ""
	}

	rootRelPath, err := fp.Rel(nsMount.root, parentMount.root)
	if err != nil {
		return ""
	}
	return fp.Join(parentMount.mountPoint, rootRelPath, nsRelPath)
}

// LoadPIDMounts retrieves the mounts for a certain PID
func (p *Resolver) LoadPIDMounts(pidPath string) *Resolver {
	p.nsMounts = getMountInfo(pidPath, p.buffer)
	return p
}
