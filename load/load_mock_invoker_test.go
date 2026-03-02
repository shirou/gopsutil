// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package load

import (
	"context"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// MockInvoker allows mocking command output for testing
type MockInvoker struct {
	commands map[string][]byte
}

// NewMockInvoker creates a new mock invoker with predefined responses
func NewMockInvoker() *MockInvoker {
	return &MockInvoker{
		commands: make(map[string][]byte),
	}
}

// SetResponse sets the response for a command
func (m *MockInvoker) SetResponse(name string, args []string, output string) {
	key := name + " " + strings.Join(args, " ")
	m.commands[key] = []byte(output)
}

// Command implements the Invoker interface
func (m *MockInvoker) Command(name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()
	return m.CommandWithContext(ctx, name, arg...)
}

// CommandWithContext implements the Invoker interface
func (m *MockInvoker) CommandWithContext(_ context.Context, name string, arg ...string) ([]byte, error) {
	key := name + " " + strings.Join(arg, " ")
	if output, ok := m.commands[key]; ok {
		return output, nil
	}
	return nil, fmt.Errorf("command not mocked: %s", key)
}

// SetupSystemMetricsMock configures the mock for system metrics testing
func (m *MockInvoker) SetupSystemMetricsMock() {
	// AIX vmstat -s output (cumulative since-boot counters)
	vmstatSOutput := `           3492560274 total address trans. faults
               116707 page ins
              3595620 page outs
                    0 paging space page ins
                    0 paging space page outs
                    0 total reclaims
            627328604 zero filled pages faults
                11232 executable filled pages faults
                    0 pages examined by clock
                    0 revolutions of the clock hand
                    0 pages freed by the clock
             46730047 backtracks
                    0 free frame waits
                    0 extend XPT waits
                26953 pending I/O waits
              3712327 start I/Os
              1823837 iodones
           5842393706 cpu context switches
             33412179 device interrupts
            357175605 software interrupts
           2684865841 decrementer interrupts
                 1106 mpc-sent interrupts
                 1106 mpc-receive interrupts
                    0 phantom interrupts
                    0 traps
          12918944607 syscalls
`
	m.SetResponse("vmstat", []string{"-s"}, vmstatSOutput)

	// AIX ps output for process state
	psOutput := `STATE
A
A
A
W
I
A
A
A
Z
A
`
	m.SetResponse("ps", []string{"-e", "-o", "state"}, psOutput)
}
