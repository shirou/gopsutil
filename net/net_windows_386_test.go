// SPDX-License-Identifier: BSD-3-Clause
//go:build windows && 386

package net

import (
	"testing"
	"unsafe"
)

func TestMibIfRow2Layout(t *testing.T) {
	var row mibIfRow2
	if offset := unsafe.Offsetof(row.TransmitLinkSpeed); offset != 1192 {
		t.Errorf("TransmitLinkSpeed offset = %d, want 1192", offset)
	}
	if size := unsafe.Sizeof(row); size != 1352 {
		t.Errorf("mibIfRow2 size = %d, want 1352", size)
	}
}
