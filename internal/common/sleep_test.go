package common_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shirou/gopsutil/internal/common"
)

func TestSleep(test *testing.T) {
	const dt = 50 * time.Millisecond
	var t = func(name string, ctx context.Context, expected error) {
		test.Run(name, func(test *testing.T) {
			var err = common.Sleep(ctx, dt)
			if !errors.Is(err, expected) {
				test.Errorf("expected %v, got %v", expected, err)
			}
		})
	}

	var ctx = context.Background()
	var canceled, cancel = context.WithCancel(ctx)
	cancel()

	t("background context", ctx, nil)
	t("canceled context", canceled, context.Canceled)
}
