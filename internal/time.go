package internal

import (
	"time"
)

// GenerateLast24HoursString generates a string that represents the last 24 hours
// Used for querying the Whoop API with a filter string
func GenerateLast24HoursString() (string, string) {
	currentTime := time.Now().UTC()

	// Calculate the start time for the last 24 hours
	startTime := currentTime.Add(-24 * time.Hour)

	// Format the start and end times
	layout := "2006-01-02T15:04:05.000Z"
	startTimeString := startTime.Format(layout)
	endTimeString := currentTime.Format(layout)

	return startTimeString, endTimeString

}
