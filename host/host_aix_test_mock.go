// SPDX-License-Identifier: BSD-3-Clause

package host

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

// SetupFDLimitsMock configures the mock for FD limits testing
func (m *MockInvoker) SetupFDLimitsMock() {
	m.SetResponse("bash", []string{"-c", "ulimit -n"}, "2048\n")
	m.SetResponse("bash", []string{"-c", "ulimit -Hn"}, "unlimited\n")
}
