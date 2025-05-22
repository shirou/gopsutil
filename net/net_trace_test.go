package net

import (
	"context"
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
	}
)

func mockConnections(_ context.Context, _ string) ([]ConnectionStat, error) {
	tIdx++
	return connections[tIdx-1], nil
}

func TestConnMapRefresh(t *testing.T) {
	defer replaceGlobalVar(&ProcConnMap, make(map[Addr]ProcNetStat))()

	assert.Empty(t, ProcConnMap)

	updateTable(context.Background(), "", mockConnections)
	assert.Len(t, ProcConnMap, 2)
	assert.Contains(t, ProcConnMap, addr1)
	assert.Contains(t, ProcConnMap, addr2)
	assert.NotContains(t, ProcConnMap, addr3)

	updateTable(context.Background(), "", mockConnections)
	assert.Len(t, ProcConnMap, 2)
	assert.Contains(t, ProcConnMap, addr1)
	assert.NotContains(t, ProcConnMap, addr2)
	assert.Contains(t, ProcConnMap, addr3)
}

func BenchmarkUpdateTable(b *testing.B) {
	defer replaceGlobalVar(&ProcConnMap, make(map[Addr]ProcNetStat))()

	b.ResetTimer()

	for range b.N {
		updateTable(context.Background(), "all", ConnectionsWithContext)
	}
}

func replaceGlobalVar[T any](target *T, replacement T) func() {
	saveVal := *target
	*target = replacement
	return func() { *target = saveVal }
}
