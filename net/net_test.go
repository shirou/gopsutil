// SPDX-License-Identifier: BSD-3-Clause
package net

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestAddrString(t *testing.T) {
	v := Addr{IP: "192.168.0.1", Port: 8000}

	s := v.String()
	assert.JSONEqf(t, `{"ip":"192.168.0.1","port":8000}`, s, "Addr string is invalid: %v", v)
}

func TestIOCountersStatString(t *testing.T) {
	v := IOCountersStat{
		Name:      "test",
		BytesSent: 100,
	}
	e := `{"name":"test","bytesSent":100,"bytesRecv":0,"packetsSent":0,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0}`
	assert.JSONEqf(t, e, v.String(), "NetIOCountersStat string is invalid: %v", v)
}

func TestProtoCountersStatString(t *testing.T) {
	v := ProtoCountersStat{
		Protocol: "tcp",
		Stats: map[string]int64{
			"MaxConn":      -1,
			"ActiveOpens":  4000,
			"PassiveOpens": 3000,
		},
	}
	e := `{"protocol":"tcp","stats":{"ActiveOpens":4000,"MaxConn":-1,"PassiveOpens":3000}}`
	assert.JSONEqf(t, e, v.String(), "NetProtoCountersStat string is invalid: %v", v)
}

func TestConnectionStatString(t *testing.T) {
	v := ConnectionStat{
		Fd:     10,
		Family: 10,
		Type:   10,
		Uids:   []int32{10, 10},
	}
	e := `{"fd":10,"family":10,"type":10,"localaddr":{"ip":"","port":0},"remoteaddr":{"ip":"","port":0},"status":"","uids":[10,10],"pid":0}`
	assert.JSONEqf(t, e, v.String(), "NetConnectionStat string is invalid: %v", v)
}

func TestIOCountersAll(t *testing.T) {
	v, err := IOCounters(false)
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "Could not get NetIOCounters: %v", err)
	per, err := IOCounters(true)
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "Could not get NetIOCounters: %v", err)
	assert.Lenf(t, v, 1, "Could not get NetIOCounters: %v", v)
	assert.Equalf(t, "all", v[0].Name, "Invalid NetIOCounters: %v", v)
	var pr uint64
	for _, p := range per {
		pr += p.PacketsRecv
	}
	// small diff is ok, compare instead of math.Abs(subtraction) with uint64
	var diff uint64
	if v[0].PacketsRecv > pr {
		diff = v[0].PacketsRecv - pr
	} else {
		diff = pr - v[0].PacketsRecv
	}
	if diff > 5 {
		if ci := os.Getenv("CI"); ci != "" {
			// This test often fails in CI. so just print even if failed.
			fmt.Printf("invalid sum value: %v, %v", v[0].PacketsRecv, pr)
		} else {
			t.Errorf("invalid sum value: %v, %v", v[0].PacketsRecv, pr)
		}
	}
}

func TestIOCountersPerNic(t *testing.T) {
	v, err := IOCounters(true)
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "Could not get NetIOCounters: %v", err)
	assert.NotEmptyf(t, v, "Could not get NetIOCounters: %v", v)
	for _, vv := range v {
		assert.NotEmptyf(t, vv.Name, "Invalid NetIOCounters: %v", vv)
	}
}

func TestGetNetIOCountersAll(t *testing.T) {
	n := []IOCountersStat{
		{
			Name:        "a",
			BytesRecv:   10,
			PacketsRecv: 10,
		},
		{
			Name:        "b",
			BytesRecv:   10,
			PacketsRecv: 10,
			Errin:       10,
		},
	}
	ret := getIOCountersAll(n)
	assert.Lenf(t, ret, 1, "invalid return count")
	assert.Equalf(t, "all", ret[0].Name, "invalid return name")
	assert.Equalf(t, uint64(20), ret[0].BytesRecv, "invalid count bytesrecv")
	assert.Equalf(t, uint64(10), ret[0].Errin, "invalid count errin")
}

func TestInterfaces(t *testing.T) {
	v, err := Interfaces()
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "Could not get NetInterfaceStat: %v", err)
	assert.NotEmptyf(t, v, "Could not get NetInterfaceStat: %v", err)
	for _, vv := range v {
		assert.NotEmptyf(t, vv.Name, "Invalid NetInterface: %v", vv)
	}
}

func TestProtoCountersStatsAll(t *testing.T) {
	v, err := ProtoCounters(nil)
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "Could not get NetProtoCounters: %v", err)
	require.NotEmptyf(t, v, "Could not get NetProtoCounters: %v", err)
	for _, vv := range v {
		assert.NotEmptyf(t, vv.Protocol, "Invalid NetProtoCountersStat: %v", vv)
		assert.NotEmptyf(t, vv.Stats, "Invalid NetProtoCountersStat: %v", vv)
	}
}

func TestProtoCountersStats(t *testing.T) {
	v, err := ProtoCounters([]string{"tcp", "ip"})
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "Could not get NetProtoCounters: %v", err)
	require.NotEmptyf(t, v, "Could not get NetProtoCounters: %v", err)
	require.Lenf(t, v, 2, "Go incorrect number of NetProtoCounters: %v", err)
	for _, vv := range v {
		if vv.Protocol != "tcp" && vv.Protocol != "ip" {
			t.Errorf("Invalid NetProtoCountersStat: %v", vv)
		}
		assert.NotEmptyf(t, vv.Stats, "Invalid NetProtoCountersStat: %v", vv)
	}
}

func TestConnections(t *testing.T) {
	if ci := os.Getenv("CI"); ci != "" { // skip if test on CI
		return
	}

	v, err := Connections("inet")
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "could not get NetConnections: %v", err)
	assert.NotEmptyf(t, v, "could not get NetConnections: %v", v)
	for _, vv := range v {
		assert.NotZerof(t, vv.Family, "invalid NetConnections: %v", vv)
	}
}

func TestFilterCounters(t *testing.T) {
	if ci := os.Getenv("CI"); ci != "" { // skip if test on CI
		return
	}

	if runtime.GOOS == "linux" {
		// some test environment has not the path.
		if !common.PathExists("/proc/sys/net/netfilter/nf_connTrackCount") {
			t.SkipNow()
		}
	}

	v, err := FilterCounters()
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "could not get NetConnections: %v", err)
	assert.NotEmptyf(t, v, "could not get NetConnections: %v", v)
	for _, vv := range v {
		assert.NotZerof(t, vv.ConnTrackMax, "nf_connTrackMax needs to be greater than zero: %v", vv)
	}
}

func TestInterfaceStatString(t *testing.T) {
	v := InterfaceStat{
		Index:        0,
		MTU:          1500,
		Name:         "eth0",
		HardwareAddr: "01:23:45:67:89:ab",
		Flags:        []string{"up", "down"},
		Addrs:        InterfaceAddrList{{Addr: "1.2.3.4"}, {Addr: "5.6.7.8"}},
	}

	s := v.String()
	assert.JSONEqf(t, `{"index":0,"mtu":1500,"name":"eth0","hardwareAddr":"01:23:45:67:89:ab","flags":["up","down"],"addrs":[{"addr":"1.2.3.4"},{"addr":"5.6.7.8"}]}`, s, "InterfaceStat string is invalid: %v", s)

	list := InterfaceStatList{v, v}
	s = list.String()
	assert.JSONEqf(t, `[{"index":0,"mtu":1500,"name":"eth0","hardwareAddr":"01:23:45:67:89:ab","flags":["up","down"],"addrs":[{"addr":"1.2.3.4"},{"addr":"5.6.7.8"}]},{"index":0,"mtu":1500,"name":"eth0","hardwareAddr":"01:23:45:67:89:ab","flags":["up","down"],"addrs":[{"addr":"1.2.3.4"},{"addr":"5.6.7.8"}]}]`, s, "InterfaceStatList string is invalid: %v", s)
}
