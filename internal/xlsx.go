// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/xuri/excelize/v2"
)

// ConvertToExcel converts the user data to a temporary CSV file
func ConvertToExcel(userData User) ([]byte, error) {

	var (
		csvData                     []byte
		timeZoneOffsetFallbackValue string
	)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("Failed to close the Excel spreadsheet", "error", err)
		}
	}()

	_, err := f.NewSheet("User Data")
	if err != nil {
		return csvData, fmt.Errorf("failed to create a new sheet - User Data: %v", err.Error())
	}

	err = f.SetSheetRow("User Data", "A1", &[]string{"UserID", "FirstName", "LastName", "Email", "Height (Meters)", "Weight (Kilograms)", "Max Heart Rate"})
	if err != nil {
		return csvData, fmt.Errorf("failed to set header row in User Data sheet: %v", err.Error())
	}
	err = f.SetSheetRow("User Data", "A2", &[]any{
		userData.UserData.UserID,
		userData.UserData.FirstName,
		userData.UserData.LastName,
		userData.UserData.Email,
		userData.UserMesaurements.HeightMeter,
		userData.UserMesaurements.WeightKilogram,
		userData.UserMesaurements.MaxHeartRate,
	})
	if err != nil {
		return csvData, fmt.Errorf("failed to set row in User Data sheet: %v", err.Error())
	}

	_, err = f.NewSheet("Sleep Data")
	if err != nil {
		return csvData, fmt.Errorf("failed to create a new sheet - Sleep Data: %v", err.Error())
	}

	sleepRowHeader := []string{
		"ID",
		"UserID",
		"Created At",
		"Updated At",
		"Start Time",
		"End Time",
		"Time Zone Offset",
		"Nap",
		"Score State",
		"Total In Bed Time (Miliseconds)",
		"Total Awake Time (Miliseconds)",
		"Total No Data Time (Miliseconds)",
		"Total Light Sleep Time (Miliseconds)",
		"Total Slow Wave Sleep Time (Miliseconds)",
		"Total REM Sleep Time (Miliseconds)",
		"Sleep Cycle Count",
		"Disturbance Count",
		"Sleep Needed Baseline (Miliseconds)",
		"Sleep Needed From Sleep Dept (Miliseconds)",
		"Sleep Needed From Recent Strain (Miliseconds)",
		"Sleep Needed From Recent Nap (Miliseconds)",
		"Respiratory Rate",
		"Sleep Performance Percentage",
		"Sleep Consistency Percentage",
		"Sleep Efficiency Percentage",
	}

	err = f.SetSheetRow("Sleep Data", "A1", &sleepRowHeader)
	if err != nil {
		return csvData, fmt.Errorf("failed to set header row in Sleep Data sheet: %v", err.Error())
	}

	rowCounter := 1
	for i := 0; i < len(userData.SleepCollection.SleepCollectionRecords); i++ {

		// Set the fallback value for the timezone offset. Required for the Recovery Data as Whoop API does not provide the timezone offset for Recovery Data
		if i == 0 {
			timeZoneOffsetFallbackValue = userData.SleepCollection.SleepCollectionRecords[i].TimezoneOffset
		}

		row := fmt.Sprintf("A%d", rowCounter+1)

		err := f.SetSheetRow("Sleep Data", row, &[]any{
			userData.SleepCollection.SleepCollectionRecords[i].ID,
			userData.SleepCollection.SleepCollectionRecords[i].UserID,
			FormatTimeWithOffset(userData.SleepCollection.SleepCollectionRecords[i].CreatedAt, userData.SleepCollection.SleepCollectionRecords[i].TimezoneOffset),
			FormatTimeWithOffset(userData.SleepCollection.SleepCollectionRecords[i].UpdatedAt, userData.SleepCollection.SleepCollectionRecords[i].TimezoneOffset),
			FormatTimeWithOffset(userData.SleepCollection.SleepCollectionRecords[i].Start, userData.SleepCollection.SleepCollectionRecords[i].TimezoneOffset),
			FormatTimeWithOffset(userData.SleepCollection.SleepCollectionRecords[i].End, userData.SleepCollection.SleepCollectionRecords[i].TimezoneOffset),
			userData.SleepCollection.SleepCollectionRecords[i].TimezoneOffset,
			userData.SleepCollection.SleepCollectionRecords[i].Nap,
			userData.SleepCollection.SleepCollectionRecords[i].ScoreState,
			// Stage Summary
			userData.SleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalInBedTimeMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalAwakeTimeMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalNoDataTimeMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalLightSleepTimeMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalSlowWaveSleepTimeMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalRemSleepTimeMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.StageSummary.SleepCycleCount,
			userData.SleepCollection.SleepCollectionRecords[i].Score.StageSummary.DisturbanceCount,
			// Sleep Needed
			userData.SleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.BaselineMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromSleepDebtMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentStrainMilli,
			userData.SleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentNapMilli,
			// Remainder of SleepCollectionRecord
			userData.SleepCollection.SleepCollectionRecords[i].Score.RespiratoryRate,
			userData.SleepCollection.SleepCollectionRecords[i].Score.SleepPerformancePercentage,
			userData.SleepCollection.SleepCollectionRecords[i].Score.SleepConsistencyPercentage,
			userData.SleepCollection.SleepCollectionRecords[i].Score.SleepEfficiencyPercentage,
		})
		if err != nil {
			return csvData, fmt.Errorf("failed to set row in Sleep Data sheet: %v", err.Error())
		}
		rowCounter++
	}

	_, err = f.NewSheet("Recovery Data")
	if err != nil {
		return csvData, fmt.Errorf("failed to create a new sheet - Recovery Data: %v", err.Error())
	}

	recoveryRowHeader := []string{
		"Cycle ID",
		"Sleep ID",
		"User ID",
		"Created At",
		"Updated At",
		"Score State",
		"User Calibrating",
		"Recovery Score",
		"Resting Heart Rate",
		"HRV RMSSD (Miliseconds)",
		"SpO2 Percentage",
		"Skin Temp (Celsius)",
	}

	err = f.SetSheetRow("Recovery Data", "A1", &recoveryRowHeader)
	if err != nil {
		return csvData, fmt.Errorf("failed to set header row in Recovery Data sheet: %v", err.Error())
	}

	rowCounter = 1
	for i := 0; i < len(userData.RecoveryCollection.RecoveryRecords); i++ {
		row := fmt.Sprintf("A%d", rowCounter+1)

		err := f.SetSheetRow("Recovery Data", row, &[]any{
			userData.RecoveryCollection.RecoveryRecords[i].CycleID,
			userData.RecoveryCollection.RecoveryRecords[i].SleepID,
			userData.RecoveryCollection.RecoveryRecords[i].UserID,
			FormatTimeWithOffset(userData.RecoveryCollection.RecoveryRecords[i].CreatedAt, timeZoneOffsetFallbackValue),
			FormatTimeWithOffset(userData.RecoveryCollection.RecoveryRecords[i].UpdatedAt, timeZoneOffsetFallbackValue),
			userData.RecoveryCollection.RecoveryRecords[i].ScoreState,
			userData.RecoveryCollection.RecoveryRecords[i].Score.UserCalibrating,
			userData.RecoveryCollection.RecoveryRecords[i].Score.RecoveryScore,
			userData.RecoveryCollection.RecoveryRecords[i].Score.RestingHeartRate,
			userData.RecoveryCollection.RecoveryRecords[i].Score.HrvRmssdMilli,
			userData.RecoveryCollection.RecoveryRecords[i].Score.Spo2Percentage,
			userData.RecoveryCollection.RecoveryRecords[i].Score.SkinTempCelsius,
		})
		if err != nil {
			return csvData, fmt.Errorf("failed to set row in Recovery Data sheet: %v", err.Error())
		}
		rowCounter++
	}

	_, err = f.NewSheet("Workout Data")
	if err != nil {
		return csvData, fmt.Errorf("failed to create a new sheet - Recovery Data: %v", err.Error())
	}

	workoutHeaders := []string{
		"ID",
		"UserID",
		"Created At",
		"Updated At",
		"Start Time",
		"End Time",
		"Time Zone Offset",
		"Sport ID",
		"Score State",
		"Strain",
		"Average Heart Rate",
		"Max Heart Rate",
		"Kilojoule",
		"Percent Recorded",
		"Distance (Meters)",
		"Altitude Gain (Meters)",
		"Altitute Change (Meters)",
		"Zone Zero Time (Miliseconds)",
		"Zone One Time (Miliseconds)",
		"Zone Two Time (Miliseconds)",
		"Zone Three Time (Miliseconds)",
		"Zone Four Time (Miliseconds)",
		"Zone Five Time (Miliseconds)",
	}

	err = f.SetSheetRow("Workout Data", "A1", &workoutHeaders)
	if err != nil {
		return csvData, fmt.Errorf("failed to set header row in Workout Data sheet: %v", err.Error())
	}
	rowCounter = 1

	for i := 0; i < len(userData.WorkoutCollection.Records); i++ {
		row := fmt.Sprintf("A%d", rowCounter+1)
		err := f.SetSheetRow("Workout Data", row, &[]any{
			userData.WorkoutCollection.Records[i].ID,
			userData.WorkoutCollection.Records[i].UserID,
			FormatTimeWithOffset(userData.WorkoutCollection.Records[i].CreatedAt, userData.WorkoutCollection.Records[i].TimezoneOffset),
			FormatTimeWithOffset(userData.WorkoutCollection.Records[i].UpdatedAt, userData.WorkoutCollection.Records[i].TimezoneOffset),
			FormatTimeWithOffset(userData.WorkoutCollection.Records[i].Start, userData.WorkoutCollection.Records[i].TimezoneOffset),
			FormatTimeWithOffset(userData.WorkoutCollection.Records[i].End, userData.WorkoutCollection.Records[i].TimezoneOffset),
			userData.WorkoutCollection.Records[i].TimezoneOffset,
			userData.WorkoutCollection.Records[i].SportID,
			userData.WorkoutCollection.Records[i].ScoreState,
			userData.WorkoutCollection.Records[i].Score.Strain,
			userData.WorkoutCollection.Records[i].Score.AverageHeartRate,
			userData.WorkoutCollection.Records[i].Score.MaxHeartRate,
			userData.WorkoutCollection.Records[i].Score.Kilojoule,
			userData.WorkoutCollection.Records[i].Score.PercentRecorded,
			userData.WorkoutCollection.Records[i].Score.DistanceMeter,
			userData.WorkoutCollection.Records[i].Score.AltitudeGainMeter,
			userData.WorkoutCollection.Records[i].Score.AltitudeChangeMeter,
			userData.WorkoutCollection.Records[i].Score.ZoneDuration.ZoneZeroMilli,
			userData.WorkoutCollection.Records[i].Score.ZoneDuration.ZoneOneMilli,
			userData.WorkoutCollection.Records[i].Score.ZoneDuration.ZoneTwoMilli,
			userData.WorkoutCollection.Records[i].Score.ZoneDuration.ZoneThreeMilli,
			userData.WorkoutCollection.Records[i].Score.ZoneDuration.ZoneFourMilli,
			userData.WorkoutCollection.Records[i].Score.ZoneDuration.ZoneFiveMilli,
		})
		if err != nil {
			return csvData, fmt.Errorf("failed to set row in Workout Data sheet: %v", err.Error())
		}
		rowCounter++
	}

	_, err = f.NewSheet("Cycle Data")
	if err != nil {
		return csvData, fmt.Errorf("failed to create a new sheet - Cycle Data: %v", err.Error())
	}

	cycleHeaders := []string{
		"ID",
		"UserID",
		"Created At",
		"Updated At",
		"Start Time",
		"End Time",
		"Time Zone Offset",
		"Score State",
		"Strain",
		"Kilojoule",
		"Average Heart Rate",
		"Max Heart Rate",
	}

	err = f.SetSheetRow("Cycle Data", "A1", &cycleHeaders)
	if err != nil {
		return csvData, fmt.Errorf("failed to set header row in Cycle Data sheet: %v", err.Error())
	}

	rowCounter = 1
	for i := 0; i < len(userData.CycleCollection.Records); i++ {
		row := fmt.Sprintf("A%d", rowCounter+1)
		err := f.SetSheetRow("Cycle Data", row, &[]any{
			userData.CycleCollection.Records[i].ID,
			userData.CycleCollection.Records[i].UserID,
			FormatTimeWithOffset(userData.CycleCollection.Records[i].CreatedAt, userData.CycleCollection.Records[i].TimezoneOffset),
			FormatTimeWithOffset(userData.CycleCollection.Records[i].UpdatedAt, userData.CycleCollection.Records[i].TimezoneOffset),
			FormatTimeWithOffset(userData.CycleCollection.Records[i].Start, userData.CycleCollection.Records[i].TimezoneOffset),
			FormatTimeWithOffset(userData.CycleCollection.Records[i].End, userData.CycleCollection.Records[i].TimezoneOffset),
			userData.CycleCollection.Records[i].TimezoneOffset,
			userData.CycleCollection.Records[i].ScoreState,
			userData.CycleCollection.Records[i].Score.Strain,
			userData.CycleCollection.Records[i].Score.Kilojoule,
			userData.CycleCollection.Records[i].Score.AverageHeartRate,
			userData.CycleCollection.Records[i].Score.MaxHeartRate,
		})
		if err != nil {
			return csvData, fmt.Errorf("failed to set row in Cycle Data sheet: %v", err.Error())
		}
		rowCounter++
	}

	err = f.DeleteSheet("Sheet1")
	if err != nil {
		return csvData, fmt.Errorf("failed to delete Sheet1. Technical error message: : %v", err.Error())
	}

	var rawDataBytes *bytes.Buffer
	rawDataBytes, err = f.WriteToBuffer()
	if err != nil {
		return csvData, fmt.Errorf("failed to write the Excel file to an internal buffer. Technical error message: : %v", err.Error())
	}
	// Write in human friendly size without using another library
	slog.Debug("Data Size", "size", formatBytes(rawDataBytes.Bytes()))
	csvData = rawDataBytes.Bytes()

	return csvData, nil
}

// formatBytes returns a human readable string of the byte size
func formatBytes(size []byte) string {
	const unit = 1024
	if len(size) < unit {
		return fmt.Sprintf("%d B", len(size))
	}
	div, exp := int64(unit), 0
	for n := len(size) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(len(size))/float64(div), "KMGTPE"[exp])
}
