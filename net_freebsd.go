// +build freebsd

package gopsutil

import (
	"errors"
)

func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	return nil, errors.New("not implemented yet")
}
