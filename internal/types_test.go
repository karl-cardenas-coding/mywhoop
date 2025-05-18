// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"
)

func TestEventString(t *testing.T) {

	var custom Event = "custom"

	tests := []struct {
		id     int
		event  Event
		result string
	}{
		{1, EventErrors, "errors"},
		{2, EventSuccess, "success"},
		{3, EventAll, "all"},
		{4, custom, "custom"},
	}

	for index, test := range tests {
		test.id = index + 1
		result := test.event.String()
		if result != test.result {
			t.Errorf("Test %d: Expected %s, got %s", test.id, test.result, result)
		}
	}
}

func TestEventFromString(t *testing.T) {

	tests := []struct {
		id     int
		event  string
		result string
	}{
		{1, "errors", "errors"},
		{2, "success", "success"},
		{3, "all", "all"},
		{4, "custom", "errors"},
	}

	for index, test := range tests {
		test.id = index + 1
		result := EventFromString(test.event)
		if result.String() != test.result {
			t.Errorf("Test %d: Expected %s, got %s", test.id, test.result, result)
		}
	}
}

func TestInitActivityIDTable(t *testing.T) {
	activityTable := InitActivityIDTable()

	// Test that the map is not empty
	if len(activityTable) == 0 {
		t.Error("Activity table should not be empty")
	}

	// Test specific known activities
	testCases := []struct {
		id     int
		name   string
		exists bool
	}{
		{-1, "Activity", true},
		{0, "Running", true},
		{1, "Cycling", true},
		{19, "Fencing", true},
		{34, "Tennis", true},
		{45, "Weightlifting", true},
		{70, "Meditation", true},
		{999, "", false}, // Non-existent ID
	}

	for _, tc := range testCases {
		name, exists := activityTable[tc.id]
		if exists != tc.exists {
			t.Errorf("Activity ID %d existence check failed: expected %v, got %v", tc.id, tc.exists, exists)
		}
		if exists && name != tc.name {
			t.Errorf("Activity ID %d name mismatch: expected %s, got %s", tc.id, tc.name, name)
		}
	}

	// Test that all values are non-empty strings
	for id, name := range activityTable {
		if name == "" {
			t.Errorf("Activity ID %d has an empty name", id)
		}
	}

	// Test that there are no duplicate names
	nameCount := make(map[string]int)
	for _, name := range activityTable {
		nameCount[name]++
		if nameCount[name] > 1 {
			t.Errorf("Duplicate activity name found: %s", name)
		}
	}
}
