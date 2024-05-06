package internal

import (
	"testing"
)

func TestGenerateLast24HoursString(t *testing.T) {

	startTimeString, endTimeString := GenerateLast24HoursString()

	if startTimeString == "" {
		t.Errorf("Expected a start time string, got: %v", startTimeString)
	}

	if endTimeString == "" {
		t.Errorf("Expected an end time string, got: %v", endTimeString)
	}

	// Check format of the strings
	if len(startTimeString) != 24 {
		t.Errorf("Expected a string of length 24, got: %v", len(startTimeString))
	}

}
