// +build plan9

package process

import (
	"syscall"
)

type Signal = syscall.Note
