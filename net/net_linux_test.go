// SPDX-License-Identifier: BSD-3-Clause
package net

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestIOCountersByFileParsing(t *testing.T) {
	// Prpare a temporary file, which will be read during the test
	tmpfile, err := os.CreateTemp("", "proc_dev_net")
	defer os.Remove(tmpfile.Name()) // clean up

	require.NoError(t, err, "Temporary file creation failed: ", err)

	cases := [4][2]string{
		{"eth0:   ", "eth1:   "},
		{"eth0:0:   ", "eth1:0:   "},
		{"eth0:", "eth1:"},
		{"eth0:0:", "eth1:0:"},
	}
	for _, testCase := range cases {
		err = tmpfile.Truncate(0)
		require.NoError(t, err, "Temporary file truncating problem: ", err)

		// Parse interface name for assertion
		interface0 := strings.TrimSpace(testCase[0])
		interface0 = interface0[:len(interface0)-1]

		interface1 := strings.TrimSpace(testCase[1])
		interface1 = interface1[:len(interface1)-1]

		// Replace the interfaces from the test case
		proc := []byte(fmt.Sprintf("Inter-|   Receive                                                |  Transmit\n face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed\n  %s1       2    3    4    5     6          7         8        9       10    11    12    13     14       15          16\n    %s100 200    300   400    500     600          700         800 900 1000    1100    1200    1300    1400       1500          1600\n", testCase[0], testCase[1]))

		// Write /proc/net/dev sample output
		_, err = tmpfile.Write(proc)
		require.NoError(t, err, "Temporary file writing failed: ", err)

		counters, err := IOCountersByFile(true, tmpfile.Name())

		require.NoError(t, err)
		assert.NotEmpty(t, counters)
		assert.Len(t, counters, 2)
		assert.Equal(t, interface0, counters[0].Name)
		assert.Equal(t, uint64(1), counters[0].BytesRecv)
		assert.Equal(t, uint64(2), counters[0].PacketsRecv)
		assert.Equal(t, uint64(3), counters[0].Errin)
		assert.Equal(t, uint64(4), counters[0].Dropin)
		assert.Equal(t, uint64(5), counters[0].Fifoin)
		assert.Equal(t, uint64(9), counters[0].BytesSent)
		assert.Equal(t, uint64(10), counters[0].PacketsSent)
		assert.Equal(t, uint64(11), counters[0].Errout)
		assert.Equal(t, uint64(12), counters[0].Dropout)
		assert.Equal(t, uint64(13), counters[0].Fifoout)
		assert.Equal(t, interface1, counters[1].Name)
		assert.Equal(t, uint64(100), counters[1].BytesRecv)
		assert.Equal(t, uint64(200), counters[1].PacketsRecv)
		assert.Equal(t, uint64(300), counters[1].Errin)
		assert.Equal(t, uint64(400), counters[1].Dropin)
		assert.Equal(t, uint64(500), counters[1].Fifoin)
		assert.Equal(t, uint64(900), counters[1].BytesSent)
		assert.Equal(t, uint64(1000), counters[1].PacketsSent)
		assert.Equal(t, uint64(1100), counters[1].Errout)
		assert.Equal(t, uint64(1200), counters[1].Dropout)
		assert.Equal(t, uint64(1300), counters[1].Fifoout)
	}

	err = tmpfile.Close()
	assert.NoError(t, err, "Temporary file closing failed: ", err)
}

func TestGetProcInodesAll(t *testing.T) {
	waitForServer := make(chan bool)
	go func(t *testing.T) { // TCP listening goroutine to have some opened inodes even in CI
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0") // dynamically get a random open port from OS
		if err != nil {
			t.Skipf("unable to resolve localhost: %v", err)
		}
		l, err := net.ListenTCP(addr.Network(), addr)
		if err != nil {
			t.Skipf("unable to listen on %v: %v", addr, err)
		}
		defer l.Close()
		waitForServer <- true
		for {
			conn, err := l.Accept()
			if err != nil {
				t.Skipf("unable to accept connection: %v", err)
			}
			defer conn.Close()
		}
	}(t)
	<-waitForServer

	ctx := context.Background()
	root := common.HostProcWithContext(ctx, "")
	v, err := getProcInodesAllWithContext(ctx, root, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, v)
}

func TestConnectionsMax(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skip CI")
	}

	maxConn := 10
	v, err := ConnectionsMax("tcp", maxConn)
	require.NoError(t, err)
	assert.NotEmpty(t, v)

	cxByPid := map[int32]int{}
	for _, cx := range v {
		if cx.Pid > 0 {
			cxByPid[cx.Pid]++
		}
	}
	for _, c := range cxByPid {
		assert.LessOrEqual(t, c, maxConn)
	}
}

type AddrTest struct {
	IP    string
	Port  int
	Error bool
}

func TestDecodeAddress(t *testing.T) {
	assert := assert.New(t)

	addr := map[string]AddrTest{
		"11111:0035": {
			Error: true,
		},
		"0100007F:BLAH": {
			Error: true,
		},
		"0085002452100113070057A13F025401:0035": {
			IP:   "2400:8500:1301:1052:a157:7:154:23f",
			Port: 53,
		},
		"00855210011307F025401:0035": {
			Error: true,
		},
	}
	if common.IsLittleEndian() {
		addr["0500000A:0016"] = AddrTest{
			IP:   "10.0.0.5",
			Port: 22,
		}
		addr["0100007F:D1C2"] = AddrTest{
			IP:   "127.0.0.1",
			Port: 53698,
		}
	} else {
		addr["0A000005:0016"] = AddrTest{
			IP:   "10.0.0.5",
			Port: 22,
		}
		addr["7F000001:D1C2"] = AddrTest{
			IP:   "127.0.0.1",
			Port: 53698,
		}
	}

	for src, dst := range addr {
		family := syscall.AF_INET
		if len(src) > 13 {
			family = syscall.AF_INET6
		}
		addr, err := decodeAddress(uint32(family), src)
		if dst.Error {
			assert.Error(err, src)
		} else {
			require.NoError(t, err, src)
			assert.Equal(dst.IP, addr.IP, src)
			assert.Equal(dst.Port, int(addr.Port), src)
		}
	}
}

func TestReverse(t *testing.T) {
	src := []byte{0x01, 0x02, 0x03}
	assert.Equal(t, []byte{0x03, 0x02, 0x01}, Reverse(src))
}

func TestConntrackStatFileParsing(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "proc_net_stat_conntrack")
	defer os.Remove(tmpfile.Name())
	require.NoError(t, err, "Temporary file creation failed: ", err)

	data := []byte(`
entries  searched found new invalid ignore delete deleteList insert insertFailed drop earlyDrop icmpError  expectNew expectCreate expectDelete searchRestart
0000007b  00000000 00000000 00000000 000b115a 00000084 00000000 00000000 00000000 00000000 00000000 00000000 00000000  00000000 00000000 00000000 0000004a
0000007b  00000000 00000000 00000000 0007eee5 00000068 00000000 00000000 00000000 00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000035
0000007b  00000000 00000000 00000000 0090346b 00000057 00000000 00000000 00000000 00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000025
0000007b  00000000 00000000 00000000 0005920f 00000069 00000000 00000000 00000000 00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000064
0000007b  00000000 00000000 00000000 000331ff 00000059 00000000 00000000 00000000 00000000 00000000 00000000 00000000  00000000 00000000 00000000 0000003b
0000007b  00000000 00000000 00000000 000314ea 00000066 00000000 00000000 00000000 00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000054
0000007b  00000000 00000000 00000000 0002b270 00000055 00000000 00000000 00000000 00000000 00000000 00000000 00000000  00000000 00000000 00000000 0000003d
0000007b  00000000 00000000 00000000 0002f67d 00000057 00000000 00000000 00000000 00000000 00000000 00000000 00000000  00000000 00000000 00000000 00000042
`)

	// Expected results
	slist := NewConntrackStatList()

	slist.Append(&ConntrackStat{
		Entries:       123,
		Searched:      0,
		Found:         0,
		New:           0,
		Invalid:       725338,
		Ignore:        132,
		Delete:        0,
		DeleteList:    0,
		Insert:        0,
		InsertFailed:  0,
		Drop:          0,
		EarlyDrop:     0,
		IcmpError:     0,
		ExpectNew:     0,
		ExpectCreate:  0,
		ExpectDelete:  0,
		SearchRestart: 74,
	})
	slist.Append(&ConntrackStat{123, 0, 0, 0, 519909, 104, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 53})

	slist.Append(&ConntrackStat{123, 0, 0, 0, 9450603, 87, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 37})
	slist.Append(&ConntrackStat{123, 0, 0, 0, 365071, 105, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100})

	slist.Append(&ConntrackStat{123, 0, 0, 0, 209407, 89, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 59})
	slist.Append(&ConntrackStat{123, 0, 0, 0, 201962, 102, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 84})

	slist.Append(&ConntrackStat{123, 0, 0, 0, 176752, 85, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 61})
	slist.Append(&ConntrackStat{123, 0, 0, 0, 194173, 87, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 66})

	// Write data to tempfile
	_, err = tmpfile.Write(data)
	require.NoError(t, err, "Temporary file writing failed: ", err)

	// Function under test
	stats, err := conntrackStatsFromFile(tmpfile.Name(), true)
	require.NoError(t, err)
	assert.Len(t, stats, 8, "Expected 8 results")

	summary := &ConntrackStat{}
	for i, exp := range slist.Items() {
		st := stats[i]

		assert.Equal(t, exp.Entries, st.Entries)
		summary.Entries += st.Entries

		assert.Equal(t, exp.Searched, st.Searched)
		summary.Searched += st.Searched

		assert.Equal(t, exp.Found, st.Found)
		summary.Found += st.Found

		assert.Equal(t, exp.New, st.New)
		summary.New += st.New

		assert.Equal(t, exp.Invalid, st.Invalid)
		summary.Invalid += st.Invalid

		assert.Equal(t, exp.Ignore, st.Ignore)
		summary.Ignore += st.Ignore

		assert.Equal(t, exp.Delete, st.Delete)
		summary.Delete += st.Delete

		assert.Equal(t, exp.DeleteList, st.DeleteList)
		summary.DeleteList += st.DeleteList

		assert.Equal(t, exp.Insert, st.Insert)
		summary.Insert += st.Insert

		assert.Equal(t, exp.InsertFailed, st.InsertFailed)
		summary.InsertFailed += st.InsertFailed

		assert.Equal(t, exp.Drop, st.Drop)
		summary.Drop += st.Drop

		assert.Equal(t, exp.EarlyDrop, st.EarlyDrop)
		summary.EarlyDrop += st.EarlyDrop

		assert.Equal(t, exp.IcmpError, st.IcmpError)
		summary.IcmpError += st.IcmpError

		assert.Equal(t, exp.ExpectNew, st.ExpectNew)
		summary.ExpectNew += st.ExpectNew

		assert.Equal(t, exp.ExpectCreate, st.ExpectCreate)
		summary.ExpectCreate += st.ExpectCreate

		assert.Equal(t, exp.ExpectDelete, st.ExpectDelete)
		summary.ExpectDelete += st.ExpectDelete

		assert.Equal(t, exp.SearchRestart, st.SearchRestart)
		summary.SearchRestart += st.SearchRestart
	}

	// Test summary grouping
	totals, err := conntrackStatsFromFile(tmpfile.Name(), false)
	require.NoError(t, err)
	for i, st := range totals {
		assert.Equal(t, summary.Entries, st.Entries)
		assert.Equal(t, summary.Searched, st.Searched)
		assert.Equal(t, summary.Found, st.Found)
		assert.Equal(t, summary.New, st.New)
		assert.Equal(t, summary.Invalid, st.Invalid)
		assert.Equal(t, summary.Ignore, st.Ignore)
		assert.Equal(t, summary.Delete, st.Delete)
		assert.Equal(t, summary.DeleteList, st.DeleteList)
		assert.Equal(t, summary.Insert, st.Insert)
		assert.Equal(t, summary.InsertFailed, st.InsertFailed)
		assert.Equal(t, summary.Drop, st.Drop)
		assert.Equal(t, summary.EarlyDrop, st.EarlyDrop)
		assert.Equal(t, summary.IcmpError, st.IcmpError)
		assert.Equal(t, summary.ExpectNew, st.ExpectNew)
		assert.Equal(t, summary.ExpectCreate, st.ExpectCreate)
		assert.Equal(t, summary.ExpectDelete, st.ExpectDelete)
		assert.Equal(t, summary.SearchRestart, st.SearchRestart)

		assert.Equal(t, 0, i) // Should only have one element
	}
}
