package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/xuri/excelize/v2"
)

func ConvertToCSV(userData User) ([]byte, error) {

	var csvData []byte
	fmt.Println("Temp Dir is: ", os.TempDir())

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

	f.SetSheetRow("User Data", "A1", &[]string{"UserID", "FirstName", "LastName", "Email", "Height (Meters)", "Weight (Kilograms)", "Max Heart Rate"})
	f.SetSheetRow("User Data", "A2", &[]any{
		userData.UserData.UserID,
		userData.UserData.FirstName,
		userData.UserData.LastName,
		userData.UserData.Email,
		userData.UserMesaurements.HeightMeter,
		userData.UserMesaurements.WeightKilogram,
		userData.UserMesaurements.MaxHeartRate,
	})

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

	f.SetSheetRow("Sleep Data", "A1", &sleepRowHeader)

	rowCounter := 1
	for i := 0; i < len(userData.SleepCollection.SleepCollectionRecords); i++ {
		row := fmt.Sprintf("A%d", rowCounter+1)

		f.SetSheetRow("Sleep Data", row, &[]any{
			userData.SleepCollection.SleepCollectionRecords[i].ID,
			userData.SleepCollection.SleepCollectionRecords[i].UserID,
			userData.SleepCollection.SleepCollectionRecords[i].CreatedAt.String(),
			userData.SleepCollection.SleepCollectionRecords[i].UpdatedAt.String(),
			userData.SleepCollection.SleepCollectionRecords[i].Start.String(),
			userData.SleepCollection.SleepCollectionRecords[i].End.String(),
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

	f.SetSheetRow("Recovery Data", "A1", &recoveryRowHeader)

	rowCounter = 1
	for i := 0; i < len(userData.RecoveryCollection.RecoveryRecords); i++ {
		row := fmt.Sprintf("A%d", rowCounter+1)

		f.SetSheetRow("Recovery Data", row, &[]any{
			userData.RecoveryCollection.RecoveryRecords[i].CycleID,
			userData.RecoveryCollection.RecoveryRecords[i].SleepID,
			userData.RecoveryCollection.RecoveryRecords[i].UserID,
			userData.RecoveryCollection.RecoveryRecords[i].CreatedAt.String(),
			userData.RecoveryCollection.RecoveryRecords[i].UpdatedAt.String(),
			userData.RecoveryCollection.RecoveryRecords[i].ScoreState,
			userData.RecoveryCollection.RecoveryRecords[i].Score.UserCalibrating,
			userData.RecoveryCollection.RecoveryRecords[i].Score.RecoveryScore,
			userData.RecoveryCollection.RecoveryRecords[i].Score.RestingHeartRate,
			userData.RecoveryCollection.RecoveryRecords[i].Score.HrvRmssdMilli,
			userData.RecoveryCollection.RecoveryRecords[i].Score.Spo2Percentage,
			userData.RecoveryCollection.RecoveryRecords[i].Score.SkinTempCelsius,
		})
		rowCounter++
	}

	rowCounter = 1
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		return csvData, fmt.Errorf("failed to delete Sheet1. Technical error message: : %v", err.Error())
	}

	err = f.SaveAs(path.Join(os.TempDir(), "mywhoop_temp.xlsx"))
	if err != nil {
		return csvData, fmt.Errorf("failed to save the Excel file. Technical error message: : %v", err.Error())
	}

	return csvData, nil
}
