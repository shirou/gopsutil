package gopsutil

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	v, err := findProcess(0)
	if err != nil {
		t.Errorf("error %v", err)
	}
	d, _ := json.Marshal(v)
	fmt.Printf("%s\n", d)
}
