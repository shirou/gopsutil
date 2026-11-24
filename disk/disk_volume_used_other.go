// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd || linux || aix

package disk

// volumeUsedBytes reports that no volume-specific used space is available, so
// Usage keeps the statfs based value.
func volumeUsedBytes(_ string) (uint64, bool) {
	return 0, false
}
