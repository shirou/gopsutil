package net

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSingleStartInvocation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer replaceGlobalVar(&ProcConnMap, nil)()

	assert.NotPanics(t, func() { StartTracing(ctx, "tcp", time.Millisecond) })
	assert.Panics(t, func() { StartTracing(ctx, "tcp", time.Millisecond) })
}

func TestGetProcConnStat(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer replaceGlobalVar(&ProcConnMap, nil)()

	StartTracing(ctx, "all", 100*time.Millisecond)
	time.Sleep(200 * time.Millisecond)

	connections := GetProcConnStat()
	assert.NotEmpty(t, connections)
}

func TestErrorReporting(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer replaceGlobalVar(&ProcConnMap, nil)()

	errChan := StartTracing(ctx, "foo", time.Millisecond)

	err := <-errChan
	assert.Error(t, err)
}

var (
	tIdx        = 0
	addr1       = Addr{"127.0.0.1", 12345}
	addr2       = Addr{"127.0.0.1", 12346}
	addr3       = Addr{"127.0.0.1", 12347}
	connections = [][]ConnectionStat{
		{ConnectionStat{Laddr: addr1, Pid: 111}, ConnectionStat{Laddr: addr2, Pid: 111}},
		{ConnectionStat{Laddr: addr1, Pid: 111}, ConnectionStat{Laddr: addr3, Pid: 111}},
		{ConnectionStat{Laddr: addr1, Pid: 111}, ConnectionStat{Laddr: addr3, Pid: 111}},
	}
)

func mockConnections(_ context.Context, _ string) ([]ConnectionStat, error) {
	ret := connections[tIdx]
	tIdx++
	return ret, nil
}

func TestConnMapRefresh(t *testing.T) {
	defer replaceGlobalVar(&ProcConnMap, make(map[Addr]*ProcNetStat))()

	expire := 20 * time.Millisecond

	assert.Empty(t, ProcConnMap)

	updateTable(context.Background(), "", mockConnections, expire)
	assert.Len(t, ProcConnMap, 2)
	assert.Contains(t, ProcConnMap, addr1)
	assert.Contains(t, ProcConnMap, addr2)
	assert.NotContains(t, ProcConnMap, addr3)

	updateTable(context.Background(), "", mockConnections, expire)
	assert.Len(t, ProcConnMap, 3)
	assert.Contains(t, ProcConnMap, addr1)
	assert.Contains(t, ProcConnMap, addr2)
	assert.Contains(t, ProcConnMap, addr3)

	time.Sleep(2 * expire)
	updateTable(context.Background(), "", mockConnections, expire)
	assert.Len(t, ProcConnMap, 2)
	assert.Contains(t, ProcConnMap, addr1)
	assert.NotContains(t, ProcConnMap, addr2)
	assert.Contains(t, ProcConnMap, addr3)
}

func TestHandlingErrorsOnRefresh(t *testing.T) {
	testErrChan := make(chan error)
	defer replaceGlobalVar(&errChan, testErrChan)()

	mockConnProvider := func(_ context.Context, _ string) ([]ConnectionStat, error) { return nil, errors.New("test") }

	go updateTable(context.Background(), "any", mockConnProvider, time.Second)

	assert.Error(t, <-testErrChan)
}

func TestGuessPidByRemote(t *testing.T) {
	testTab := make(map[Addr]*ProcNetStat)
	rAddr := Addr{IP: "104.18.138.67", Port: 443}
	lAddr := Addr{IP: "192.168.0.235", Port: 20781}
	lAddr2 := Addr{IP: "192.168.0.235", Port: 21326}
	testTab[lAddr] = &ProcNetStat{Pid: -1, RemoteAddr: rAddr}
	defer replaceGlobalVar(&ProcConnMap, testTab)()

	mockConnections := []ConnectionStat{{Pid: 123, Raddr: rAddr, Laddr: lAddr2}}

	assert.EqualValues(t, -1, testTab[lAddr].Pid)

	guessPidByRemote(mockConnections)

	assert.EqualValues(t, 123, testTab[lAddr].Pid)
}

func BenchmarkUpdateTable(b *testing.B) {
	defer replaceGlobalVar(&ProcConnMap, make(map[Addr]*ProcNetStat))()

	b.ResetTimer()

	for range b.N {
		updateTable(context.Background(), "all", ConnectionsWithContext, time.Second)
	}
}

func BenchmarkGetProcConnStat(b *testing.B) {
	defer replaceGlobalVar(&ProcConnMap, make(map[Addr]*ProcNetStat))()

	b.ResetTimer()

	for range b.N {
		GetProcConnStat()
	}
}

func replaceGlobalVar[T any](target *T, replacement T) func() {
	saveVal := *target
	*target = replacement
	return func() { *target = saveVal }
}

func toSlice[T any](vals ...T) []T {
	return vals
}
