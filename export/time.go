package export

import (
	"time"
)

// getCurrentDate returns the current date in the format "YYYY-MM-DD"
func getCurrentDate() string {
	currentDate := time.Now().Format("2006_01_02")
	return currentDate
}
