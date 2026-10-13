package battery

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestBatteryWithContext(t *testing.T) {
	var w *common.Warnings
	results, err := BatteryWithContext(context.Background())
	fmt.Printf("Battery: %+v\n", results)
	if errors.As(err, &w) {
		w.Verbose = true
		t.Logf("%v", w)
	}

	//if err != nil {
	//	fmt.Println(err)
	//}
	//for _, item := range list {
	//	fmt.Println(item)
	//}
}
