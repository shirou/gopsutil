// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeInvoker returns canned outputs per command for testing.
type fakeInvoker struct {
	outputs map[string][]byte
}

func (i fakeInvoker) Command(name string, arg ...string) ([]byte, error) {
	return i.CommandWithContext(context.Background(), name, arg...)
}

func (i fakeInvoker) CommandWithContext(_ context.Context, name string, arg ...string) ([]byte, error) {
	key := name
	if len(arg) > 0 {
		key += " " + strings.Join(arg, " ")
	}
	if out, ok := i.outputs[key]; ok {
		return out, nil
	}
	return nil, fmt.Errorf("unexpected command: %s", key)
}

func setTestInvoker(t *testing.T, outputs map[string][]byte) {
	testInvoker = fakeInvoker{outputs: outputs}
	t.Cleanup(func() { testInvoker = nil })
}

func TestBootTimeWithContext(t *testing.T) {
	// This is a wrapper function that delegates to common.BootTimeWithContext
	// Actual implementation testing is done in common_aix_test.go
	bootTime, err := BootTimeWithContext(context.TODO())
	require.NoError(t, err)
	assert.Positive(t, bootTime)
}

func TestUptimeWithContext(t *testing.T) {
	// This is a wrapper function that delegates to common.UptimeWithContext
	// Actual implementation testing is done in common_aix_test.go
	uptime, err := UptimeWithContext(context.TODO())
	require.NoError(t, err)
	assert.Positive(t, uptime)
}

func TestPlatformInformationWithContext(t *testing.T) {
	setTestInvoker(t, map[string][]byte{
		"uname -s":           []byte("AIX\n"),
		"lslpp -Lqc bos.rte": []byte("bos:bos.rte:7.3.2.2: : :C:F:Base Operating System Runtime: : : : : : :0:0:/:2419\n"),
	})

	platform, family, version, err := PlatformInformationWithContext(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "AIX", platform)
	assert.Equal(t, "AIX", family)
	assert.Equal(t, "7.3.0.0", version)
}

func TestKernelVersionWithContext(t *testing.T) {
	setTestInvoker(t, map[string][]byte{
		"lslpp -Lqc bos.rte": []byte("bos:bos.rte:7.3.2.2: : :C:F:Base Operating System Runtime: : : : : : :0:0:/:2419\n"),
	})

	version, err := KernelVersionWithContext(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "7300-02-02-2419", version)
}

func TestPlatformInformationWithContextInvalidOutput(t *testing.T) {
	// Short/missing fields in lslpp output should result in an error
	setTestInvoker(t, map[string][]byte{
		"uname -s":           []byte("AIX\n"),
		"lslpp -Lqc bos.rte": []byte("bos:bos.rte:7.3.2.2\n"),
	})

	_, _, _, err := PlatformInformationWithContext(context.Background())
	require.Error(t, err)

	_, err = KernelVersionWithContext(context.Background())
	require.Error(t, err)
}

func TestKernelVersionWithContextInvalidLevel(t *testing.T) {
	// A non-numeric level should result in an error
	setTestInvoker(t, map[string][]byte{
		"lslpp -Lqc bos.rte": []byte("bos:bos.rte:unknown: : :C:F:Base Operating System Runtime: : : : : : :0:0:/:2419\n"),
	})

	_, err := KernelVersionWithContext(context.Background())
	require.Error(t, err)
}
