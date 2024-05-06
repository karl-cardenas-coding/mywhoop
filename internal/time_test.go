package internal

import (
	"regexp"
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

func TestGetCurrentDate(t *testing.T) {
	datePattern := `^\d{4}_\d{2}_\d{2}$`
	currentDate := GetCurrentDate()

	match, err := regexp.MatchString(datePattern, currentDate)
	if err != nil {
		t.Fatalf("Error while matching date pattern: %v", err)
	}

	if !match {
		t.Errorf("Expected date format YYYY-MM-DD, but got %s", currentDate)
	}
}
