// +build linux

package so

import (
	"bufio"
	"path/filepath"
	"strings"
)

type pathResolver struct {
	root *mountInfo
}

func newPathResolver(procRoot string, b *bufio.Reader) *pathResolver {
	initPIDPath := filepath.Join(procRoot, "1")
	return &pathResolver{
		root: getMountInfo(initPIDPath, b),
	}
}

// Resolve a path from a potentially namespaced process to the host path
func (p *pathResolver) Resolve(path string, nsMounts *mountInfo) string {
	if p.root == nil || nsMounts == nil {
		return ""
	}

	nsMount := nsMounts.GetMount(path)
	if nsMount == nil {
		return ""
	}
	nsRelPath, err := filepath.Rel(nsMount.mountPoint, path)
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

	rootRelPath, err := filepath.Rel(nsMount.root, parentMount.root)
	if err != nil {
		return ""
	}
	return filepath.Join(parentMount.mountPoint, rootRelPath, nsRelPath)
}
