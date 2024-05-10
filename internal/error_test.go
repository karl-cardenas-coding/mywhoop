// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"testing"
)

func TestLogError(t *testing.T) {
	// Test case 1: Error with file, line, and function information
	err := errors.New("Something went wrong")
	LogError(err)

	// Test case 2: Error without file, line, and function information
	err = errors.New("Another error occurred")
	LogError(err)
}
