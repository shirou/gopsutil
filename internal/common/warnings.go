// SPDX-License-Identifier: BSD-3-Clause
package common

import "fmt"

const maxWarnings = 100 // An arbitrary limit to avoid excessive memory usage, it has no sense to store thousands of errors

type Warnings struct {
	List    []error
	Verbose bool
}

func (w *Warnings) Add(err error) {
	if len(w.List) >= maxWarnings {
		return
	}
	w.List = append(w.List, err)
}

func (w *Warnings) Reference() error {
	if len(w.List) > 0 {
		return w
	}
	return nil
}

func (w *Warnings) Error() string {
	if w.Verbose {
		str := ""
		for i, e := range w.List {
			str += fmt.Sprintf("\tError %d: %s\n", i, e.Error())
		}
		return str
	}
	return fmt.Sprintf("Number of warnings: %v", len(w.List))
}
