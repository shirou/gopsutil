// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd

package load

func getForkStat() (forkstat, error) {
	return forkstat{}, nil
}
