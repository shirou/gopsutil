package net

import (
	"fmt"
	"syscall"
	"testing"

	"github.com/shirou/gopsutil/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestGetProcInodes(t *testing.T) {
	root := common.HostProc("")

	// /proc/19957/fd

	v, err := getProcInodes(root, 19957)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(v)
}

type AddrTest struct {
	IP    string
	Port  int
	Error bool
}

func TestDecodeAddress(t *testing.T) {
	assert := assert.New(t)

	addr := map[string]AddrTest{
		"0500000A:0016": AddrTest{
			IP:   "10.0.0.5",
			Port: 22,
		},
		"0100007F:D1C2": AddrTest{
			IP:   "127.0.0.1",
			Port: 53698,
		},
		"11111:0035": AddrTest{
			Error: true,
		},
		"0100007F:BLAH": AddrTest{
			Error: true,
		},
		"0085002452100113070057A13F025401:0035": AddrTest{
			IP:   "2400:8500:1301:1052:a157:7:154:23f",
			Port: 53,
		},
		"00855210011307F025401:0035": AddrTest{
			Error: true,
		},
	}

	for src, dst := range addr {
		family := syscall.AF_INET
		if len(src) > 13 {
			family = syscall.AF_INET6
		}
		ip, port, err := decodeAddress(family, src)
		if dst.Error {
			assert.NotNil(err, src)
		} else {
			assert.Nil(err, src)
			assert.Equal(dst.IP, ip.String(), src)
			assert.Equal(dst.Port, port, src)
		}
	}
}
