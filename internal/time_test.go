// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"regexp"
	"testing"
	"time"
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

func TestFormatTimeWithOffset(t *testing.T) {
	// Define a common time to use for testing
	baseTime := time.Date(2024, time.July, 28, 15, 43, 0, 0, time.UTC)

	tests := []struct {
		name        string
		offsetStr   string
		expected    string
		description string
	}{
		{
			name:        "UTC offset -07:00",
			offsetStr:   "-07:00",
			expected:    "28/07/2024 08:43:00",
			description: "Tests the case where the time should be adjusted by subtracting 7 hours from UTC.",
		},
		{
			name:        "UTC offset +05:30",
			offsetStr:   "+05:30",
			expected:    "28/07/2024 21:13:00",
			description: "Tests the case where the time should be adjusted by adding 5 hours and 30 minutes to UTC.",
		},
		{
			name:        "UTC offset +00:00",
			offsetStr:   "+00:00",
			expected:    "28/07/2024 15:43:00",
			description: "Tests the case where there is no offset, and the time should remain the same.",
		},
		{
			name:        "Invalid offset",
			offsetStr:   "invalid",
			expected:    "28/07/2024 15:43:00",
			description: "Tests the case where the offset string is invalid, and the function should return the original time without adjustment.",
		},
		{
			name:        "Invalid hour format",
			offsetStr:   "-ab:00",
			expected:    "28/07/2024 15:43:00",
			description: "Tests the case where the hour part of the offset string is not a valid number, and the function should return the original time without adjustment.",
		},
		{
			name:        "Invalid minute format",
			offsetStr:   "-07:xy",
			expected:    "28/07/2024 15:43:00",
			description: "Tests the case where the minute part of the offset string is not a valid number, and the function should return the original time without adjustment.",
		},
		{
			name:        "Empty Offset",
			offsetStr:   "",
			expected:    "28/07/2024 15:43:00",
			description: "Tests the case where the minute part of the offset string is not a valid number, and the function should return the original time without adjustment.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description) // Log the description to understand the purpose of the test
			result := FormatTimeWithOffset(baseTime, tt.offsetStr)
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}
