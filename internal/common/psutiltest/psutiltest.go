// SPDX-License-Identifier: BSD-3-Clause

// Package psutiltest provides helpers for tests that compare gopsutil
// results against Python psutil, the reference implementation.
//
// Tests are skipped when no python interpreter with psutil is
// available — Eval enforces this on its own, and RequirePsutil does the
// same up front so a test can skip before doing any gopsutil work.
// Setting the GOPSUTIL_PSUTIL_TEST environment variable turns those
// skips into failures (strict mode).
package psutiltest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// EnvStrict is the environment variable that, when set to a non-empty
// value, turns "python/psutil not available" skips into test failures.
// Intended for CI environments where psutil is expected to be installed.
const EnvStrict = "GOPSUTIL_PSUTIL_TEST"

// EnvPython is the environment variable that overrides python
// interpreter discovery with an explicit path.
const EnvPython = "GOPSUTIL_PSUTIL_PYTHON"

const evalTimeout = 30 * time.Second

// pyWrapper converts the result of the python expression given in
// sys.argv[1] (psutil namedtuples, dicts, lists or scalars) to JSON on
// stdout.
const pyWrapper = `
import json, sys
import psutil

def conv(o):
    if hasattr(o, "_asdict"):
        o = o._asdict()
    if isinstance(o, dict):
        return {str(k): conv(v) for k, v in o.items()}
    if isinstance(o, (list, tuple, set, frozenset)):
        return [conv(v) for v in o]
    if isinstance(o, (bool, int, float, str)) or o is None:
        return o
    return str(o)

print(json.dumps(conv(eval(sys.argv[1]))))
`

type pythonInfo struct {
	path    string
	version string
	err     error
}

var findPython = sync.OnceValue(func() pythonInfo {
	// "python" after "python3" covers Windows, where the interpreter is
	// usually named python.exe and "python3" may resolve to the Microsoft
	// Store shim, which fails the import probe below.
	candidates := []string{os.Getenv(EnvPython), "python3", "python"}
	var reasons []string
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		path, err := exec.LookPath(candidate)
		if err != nil {
			reasons = append(reasons, err.Error())
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), evalTimeout)
		out, err := exec.CommandContext(ctx, path, "-c", "import psutil, sys; sys.stdout.write(psutil.__version__)").Output()
		cancel()
		if err != nil {
			reasons = append(reasons, fmt.Sprintf("%s: %s", path, execErrDetail(err)))
			continue
		}
		return pythonInfo{path: path, version: strings.TrimSpace(string(out))}
	}
	return pythonInfo{err: fmt.Errorf("python with psutil not found: %s", strings.Join(reasons, "; "))}
})

var logVersionOnce sync.Once

// RequirePsutil skips the test (or fails it, in strict mode) unless a
// python interpreter that can import psutil is available, and logs the
// interpreter path and psutil version once per test binary. Calling it
// at the start of a test is recommended so the test skips before doing
// any gopsutil work, but Eval enforces the same contract by itself.
func RequirePsutil(tb testing.TB) {
	tb.Helper()
	info := requirePython(tb)
	logVersionOnce.Do(func() {
		tb.Logf("using %s (psutil %s)", info.path, info.version)
	})
}

// requirePython enforces the availability contract: skip by default,
// fail when strict mode (EnvStrict) is enabled.
func requirePython(tb testing.TB) pythonInfo {
	tb.Helper()
	info := findPython()
	if info.err != nil {
		if os.Getenv(EnvStrict) != "" {
			tb.Fatalf("%s is set but %v", EnvStrict, info.err)
		}
		tb.Skipf("%v", info.err)
	}
	return info
}

// Eval evaluates a python expression with psutil pre-imported, converts
// the result to JSON and decodes it into a T. It skips the test (or
// fails it, in strict mode) when no python with psutil is available;
// any other failure is returned as an error so retrying callers
// (require.EventuallyWithT callbacks) can handle it.
func Eval[T any](tb testing.TB, expr string) (T, error) {
	tb.Helper()
	var out T
	info := requirePython(tb)
	ctx, cancel := context.WithTimeout(context.Background(), evalTimeout)
	defer cancel()
	// #nosec G204 -- test-only helper; the interpreter is probed above and
	// expr always comes from test code constants.
	raw, err := exec.CommandContext(ctx, info.path, "-c", pyWrapper, expr).Output()
	if err != nil {
		return out, fmt.Errorf("eval %q: %s", expr, execErrDetail(err))
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("eval %q: unmarshal %q: %w", expr, raw, err)
	}
	return out, nil
}

func execErrDetail(err error) string {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return fmt.Sprintf("%v: %s", err, exitErr.Stderr)
	}
	return err.Error()
}
