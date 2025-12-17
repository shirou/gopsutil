// SPDX-License-Identifier: BSD-3-Clause

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
func (m *MockInvoker) CommandWithContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	key := name + " " + strings.Join(arg, " ")
	if output, ok := m.commands[key]; ok {
		return output, nil
	}
	return nil, fmt.Errorf("command not mocked: %s", key)
}

// SetupSystemMetricsMock configures the mock for system metrics testing
func (m *MockInvoker) SetupSystemMetricsMock() {
	// AIX vmstat output
	vmstatOutput := `System configuration: lcpu=8 mem=4096MB ent=0.20

kthr    memory              page              faults              cpu          
----- ----------- ------------------------ ------------ -----------------------
 r  b   avm   fre  re  pi  po  fr   sr  cy  in   sy  cs us sy id wa    pc    ec
 0  0 389541 553213   0   0   0   0    0   0   9 1083 669  1  1 98  0  0.01   6.5
`
	m.SetResponse("vmstat", []string{"1", "1"}, vmstatOutput)

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
