// +build nacl

package cpu

import (
	"time"
)

func Times(percpu bool) ([]TimesStat, error) {
	return []TimesStat{}, nil
}

func Percent(interval time.Duration, percpu bool) ([]float64, error) {
	return []float64{}, nil
}