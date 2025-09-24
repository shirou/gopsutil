// SPDX-License-Identifier: BSD-3-Clause
package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWarnings_AddAndReference(t *testing.T) {
	w := &Warnings{}
	require.NoError(t, w.Reference(), "Expected nil reference for empty warnings")
	w.Add(errors.New("first error"))
	require.Error(t, w.Reference(), "Expected non-nil reference after adding error")
	assert.Len(t, w.List, 1, "Expected 1 warning")
}

func TestWarnings_MaxLimit(t *testing.T) {
	w := &Warnings{}
	for i := range maxWarnings + 10 {
		w.Add(fmt.Errorf("error %d", i))
	}
	assert.Len(t, w.List, maxWarnings, "Expected maxWarnings warnings")
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
	for i := range maxWarnings + 1 {
		w.Add(fmt.Errorf("err%d", i))
	}
	msg := w.Error()
	assert.Contains(t, msg, tooManyErrorsMessage, "Expected too many errors message in verbose output")
	w.Verbose = false
	msg = w.Error()
	expected := fmt.Sprintf("%s > %v - %s", numberOfWarningsMessage, maxWarnings, tooManyErrorsMessage)
	assert.Equal(t, expected, msg, "Expected too many errors message in non-verbose output")
}
