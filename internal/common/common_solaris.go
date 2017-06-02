package common

import (
	"os/exec"
	"sync"
)

var (
	isainfoLock sync.RWMutex
	isainfoPath string

	kstatLock sync.RWMutex
	kstatPath string

	psrinfoLock sync.RWMutex
	psrinfoPath string

	// inZone is true when it has been detected that a process is running inside
	// of an Illumos zone (see zones(5)).
	inZoneLock sync.RWMutex
	inZone     *bool
)

// InZone returns a bool depending on whether or not the process is running in a
// zone or not.
func InZone() (bool, error) {
	inZoneLock.RLock()
	if inZone != nil {
		b := *inZone
		inZoneLock.RUnlock()
		return b, nil
	}
	inZoneLock.RUnlock()

	inZoneLock.Lock()
	defer inZoneLock.Unlock()

	zonenamePath, err := exec.LookPath("/usr/bin/zonename")
	if err != nil {
		return false, err
	}

	invoke := Invoke{}
	out, err := invoke.Command(zonenamePath)
	if err != nil {
		return false, err
	}

	var b bool
	if string(out) != "global" {
		b = true
	}
	inZone = &b

	return b, nil
}

// ISAInfoPath returns the path to isainfo(1).  ISAInfoPath() memoizes the
// stat(2) call.
func ISAInfoPath() (string, error) {
	isainfoLock.RLock()
	if len(isainfoPath) > 0 {
		isainfoLock.RUnlock()
		return isainfoPath, nil
	}
	isainfoLock.RUnlock()

	isainfoLock.Lock()
	defer isainfoLock.Unlock()

	var err error
	isainfoPath, err = exec.LookPath("/usr/bin/isainfo")
	if err != nil {
		return "", err
	}

	return isainfoPath, nil
}

// KStatPath returns the path to kstat(1M).  KStatPath() memoizes the stat(2)
// call.
func KStatPath() (string, error) {
	kstatLock.RLock()
	if len(kstatPath) > 0 {
		kstatLock.RUnlock()
		return kstatPath, nil
	}
	kstatLock.RUnlock()

	kstatLock.Lock()
	defer kstatLock.Unlock()

	var err error
	kstatPath, err = exec.LookPath("/usr/bin/kstat")
	if err != nil {
		return "", err
	}

	return kstatPath, nil
}

// ProcessorInfo returns the path to psrinfo(1M)
func ProcessorInfoPath() (string, error) {
	psrinfoLock.RLock()
	if len(psrinfoPath) > 0 {
		psrinfoLock.RUnlock()
		return psrinfoPath, nil
	}
	psrinfoLock.RUnlock()

	psrinfoLock.Lock()
	defer psrinfoLock.Unlock()

	var err error
	psrinfoPath, err = exec.LookPath("/usr/sbin/psrinfo")
	if err != nil {
		return "", err
	}

	return psrinfoPath, nil
}
