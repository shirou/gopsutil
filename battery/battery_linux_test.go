package battery

import (
	"context"
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestBatteryWithContext(t *testing.T) {
	var w *common.Warnings
	_, err := BatteryWithContext(context.Background())
	if err != nil {
		//If its just warnings, all good, otherwise, doesn't pass test.
		if errors.As(err, &w) {
			w.Verbose = true
			t.Logf("%v", w)
		} else {
			t.Fatalf("BatteryWithContext failed: %v", err)
		}
	}
}
