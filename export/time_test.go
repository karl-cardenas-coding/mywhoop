// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"testing"
	"time"
)

func TestGetCurrentDate(t *testing.T) {

	expected := time.Now().Format("2006_01_02")

	currentDate := getCurrentDate()
	if currentDate != expected {
		t.Errorf("Expected %s, got %s", expected, currentDate)
	}
}
