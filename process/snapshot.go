package process

import (
	"time"
)

// CommonSnapshot is process information at a moment in time
type CommonSnapshot struct {
	*Process
	Timestamp time.Time
}

// NewCommonSnapshot - returns a new common snapshot
func NewCommonSnapshot(pid int32) (s *CommonSnapshot, err error) {
	process, err := NewProcess(pid)
	s = &CommonSnapshot{
		Process:   process,
		Timestamp: time.Now(),
	}
	return
}
