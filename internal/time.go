// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"log/slog"
	"strconv"
	"time"
)

// GenerateLast24HoursString generates a string that represents the last 24 hours
// Used for querying the Whoop API with a filter string
func GenerateLast24HoursString() (string, string) {
	currentTime := time.Now().Local().UTC()

	// Calculate the start time for the last 24 hours
	startTime := currentTime.Add(-24 * time.Hour)

	// Format the start and end times
	layout := "2006-01-02T15:04:05.000Z"
	startTimeString := startTime.Format(layout)
	endTimeString := currentTime.Format(layout)

	slog.Debug("Time Filters", "start", startTimeString, "end", endTimeString)

	return startTimeString, endTimeString

}

// getCurrentDate returns the current date in the format "YYYY-MM-DD"
func GetCurrentDate() string {
	currentDate := time.Now().Format("2006_01_02")
	return currentDate
}

// FormatTimeWithOffset formats a time.Time object with a given offset string
// The offset string should be in the format "+HH:MM" or "-HH:MM"
func FormatTimeWithOffset(t time.Time, offsetStr string) string {

	if offsetStr == "" {
		return t.Format("02/01/2006 15:04:05")
	}

	// Attempt to parse the offset string, e.g., "-07:00"
	sign := offsetStr[0:1] // "+" or "-"
	hours, err := strconv.Atoi(offsetStr[1:3])
	if err != nil {
		return t.Format("02/01/2006 15:04:05")
	}
	minutes, err := strconv.Atoi(offsetStr[4:6])
	if err != nil {
		return t.Format("02/01/2006 15:04:05")
	}

	offset := (hours * 60 * 60) + (minutes * 60)
	if sign == "-" {
		offset = -offset
	}

	location := time.FixedZone("CustomZone", offset)
	t = t.In(location)

	return t.Format("02/01/2006 15:04:05")
}
