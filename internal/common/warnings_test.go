// SPDX-License-Identifier: BSD-3-Clause
package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWarnings_AddAndReference(t *testing.T) {
	w := &Warnings{}
	assert.Nil(t, w.Reference(), "Expected nil reference for empty warnings")
	w.Add(errors.New("first error"))
	assert.NotNil(t, w.Reference(), "Expected non-nil reference after adding error")
	assert.Equal(t, 1, len(w.List), "Expected 1 warning")
}

func TestWarnings_MaxLimit(t *testing.T) {
	w := &Warnings{}
	for i := 0; i < maxWarnings+10; i++ {
		w.Add(fmt.Errorf("error %d", i))
	}
	assert.Equal(t, maxWarnings, len(w.List), "Expected maxWarnings warnings")
	assert.True(t, w.tooManyErrors, "Expected tooManyErrors to be true after exceeding maxWarnings")
}

func TestWarnings_ErrorVerbose(t *testing.T) {
	w := &Warnings{Verbose: true}
	w.Add(errors.New("err1"))
	w.Add(errors.New("err2"))
	msg := w.Error()
	assert.NotEmpty(t, msg, "Expected verbose error string")
	assert.NotEqual(t, tooManyErrorsMessage, msg, "Expected verbose error string, not tooManyErrorsMessage")
	assert.Contains(t, msg, "Error 0: err1", "Verbose error string missing expected error 0")
	assert.Contains(t, msg, "Error 1: err2", "Verbose error string missing expected error 1")
}

func TestWarnings_ErrorNonVerbose(t *testing.T) {
	w := &Warnings{}
	w.Add(errors.New("err1"))
	msg := w.Error()
	assert.Equal(t, fmt.Sprintf("Number of warnings: %v", len(w.List)), msg, "Expected non-verbose error string")
}

func TestWarnings_ErrorTooMany(t *testing.T) {
	w := &Warnings{Verbose: true}
	for i := 0; i < maxWarnings+1; i++ {
		w.Add(fmt.Errorf("err%d", i))
	}
	msg := w.Error()
	assert.Contains(t, msg, tooManyErrorsMessage, "Expected too many errors message in verbose output")
	w.Verbose = false
	msg = w.Error()
	assert.Equal(t, tooManyErrorsMessage, msg, "Expected too many errors message in non-verbose output")
}
