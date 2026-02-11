// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestConvertToSQLite(t *testing.T) {
	timestamp := time.Date(2024, time.August, 15, 12, 0, 0, 0, time.UTC)

	userData := User{
		UserData: UserData{
			UserID:    111111111,
			Email:     "john.doe@example.com",
			FirstName: "John",
			LastName:  "Doe",
		},
		UserMeasurements: UserMeasurements{
			HeightMeter:    1.8,
			WeightKilogram: 80.1,
			MaxHeartRate:   180,
		},
		SleepCollection: SleepCollection{
			SleepCollectionRecords: []SleepCollectionRecords{
				{
					ID:             "sleep-1",
					UserID:         111111111,
					CreatedAt:      timestamp,
					UpdatedAt:      timestamp,
					Start:          timestamp,
					End:            timestamp.Add(8 * time.Hour),
					TimezoneOffset: "-07:00",
					Nap:            false,
					ScoreState:     "SCORED",
					Score: Score{
						StageSummary: StageSummary{
							TotalInBedTimeMilli:         1000,
							TotalAwakeTimeMilli:         2000,
							TotalNoDataTimeMilli:        3000,
							TotalLightSleepTimeMilli:    4000,
							TotalSlowWaveSleepTimeMilli: 5000,
							TotalRemSleepTimeMilli:      6000,
							SleepCycleCount:             7,
							DisturbanceCount:            8,
						},
						SleepNeeded: SleepNeeded{
							BaselineMilli:             9,
							NeedFromSleepDebtMilli:    10,
							NeedFromRecentStrainMilli: 11,
							NeedFromRecentNapMilli:    12,
						},
						RespiratoryRate:            13.1,
						SleepPerformancePercentage: 14.2,
						SleepConsistencyPercentage: 15.3,
						SleepEfficiencyPercentage:  16.4,
					},
				},
			},
		},
		RecoveryCollection: RecoveryCollection{
			RecoveryRecords: []RecoveryRecords{
				{
					CycleID:    12,
					SleepID:    "sleep-1",
					UserID:     111111111,
					CreatedAt:  timestamp,
					UpdatedAt:  timestamp,
					ScoreState: "SCORED",
					Score: RecoveryScore{
						UserCalibrating:  false,
						RecoveryScore:    80,
						RestingHeartRate: 52,
						HrvRmssdMilli:    90.1,
						Spo2Percentage:   98.6,
						SkinTempCelsius:  36.9,
					},
				},
			},
		},
		WorkoutCollection: WorkoutCollection{
			Records: []WorkoutRecords{
				{
					ID:             "workout-1",
					UserID:         111111111,
					CreatedAt:      timestamp,
					UpdatedAt:      timestamp,
					Start:          timestamp,
					End:            timestamp.Add(45 * time.Minute),
					TimezoneOffset: "-07:00",
					SportID:        44,
					SportName:      "Running",
					ScoreState:     "SCORED",
					Score: WorkoutScore{
						Strain:              7.7,
						AverageHeartRate:    140,
						MaxHeartRate:        188,
						Kilojoule:           500,
						PercentRecorded:     99.9,
						DistanceMeter:       10000,
						AltitudeGainMeter:   220,
						AltitudeChangeMeter: 120,
						ZoneDuration: ZoneDuration{
							ZoneZeroMilli:  1,
							ZoneOneMilli:   2,
							ZoneTwoMilli:   3,
							ZoneThreeMilli: 4,
							ZoneFourMilli:  5,
							ZoneFiveMilli:  6,
						},
					},
				},
			},
		},
		CycleCollection: CycleCollection{
			Records: []CycleRecords{
				{
					ID:             99,
					UserID:         111111111,
					CreatedAt:      timestamp,
					UpdatedAt:      timestamp,
					Start:          timestamp,
					End:            timestamp.Add(24 * time.Hour),
					TimezoneOffset: "-07:00",
					ScoreState:     "SCORED",
					Score: CycleScore{
						Strain:           12.4,
						Kilojoule:        700,
						AverageHeartRate: 133,
						MaxHeartRate:     170,
					},
				},
			},
		},
	}

	sqliteData, err := ConvertToSQLite(userData)
	if err != nil {
		t.Fatalf("ConvertToSQLite returned an error: %v", err)
	}

	if len(sqliteData) < 16 || string(sqliteData[:16]) != "SQLite format 3\x00" {
		t.Fatalf("ConvertToSQLite did not produce a valid sqlite file header")
	}

	db := openSQLiteFromBytes(t, sqliteData)

	assertSQLiteTableCount(t, db, "user_data", 1)
	assertSQLiteTableCount(t, db, "user_measurements", 1)
	assertSQLiteTableCount(t, db, "sleep_records", 1)
	assertSQLiteTableCount(t, db, "recovery_records", 1)
	assertSQLiteTableCount(t, db, "workout_records", 1)
	assertSQLiteTableCount(t, db, "cycle_records", 1)

	var email string
	if err := db.QueryRow(`SELECT email FROM user_data WHERE user_id = ?`, userData.UserData.UserID).Scan(&email); err != nil {
		t.Fatalf("failed to query user_data table: %v", err)
	}
	if email != userData.UserData.Email {
		t.Fatalf("unexpected email stored in sqlite file. expected %s, got %s", userData.UserData.Email, email)
	}

	var scoreState string
	if err := db.QueryRow(`SELECT score_state FROM workout_records WHERE id = ?`, "workout-1").Scan(&scoreState); err != nil {
		t.Fatalf("failed to query workout_records table: %v", err)
	}
	if scoreState != "SCORED" {
		t.Fatalf("unexpected score state in workout_records. expected SCORED, got %s", scoreState)
	}

	var exportRuns int
	if err := db.QueryRow(`SELECT COUNT(*) FROM export_runs`).Scan(&exportRuns); err != nil {
		t.Fatalf("failed to query export_runs table: %v", err)
	}
	if exportRuns != 1 {
		t.Fatalf("unexpected export run count. expected 1, got %d", exportRuns)
	}
}

func TestConvertToSQLiteEmptyCollections(t *testing.T) {
	userData := User{
		UserData: UserData{
			UserID:    222222222,
			Email:     "jane.doe@example.com",
			FirstName: "Jane",
			LastName:  "Doe",
		},
		UserMeasurements: UserMeasurements{
			HeightMeter:    1.7,
			WeightKilogram: 65.2,
			MaxHeartRate:   175,
		},
	}

	sqliteData, err := ConvertToSQLite(userData)
	if err != nil {
		t.Fatalf("ConvertToSQLite returned an error: %v", err)
	}

	db := openSQLiteFromBytes(t, sqliteData)

	assertSQLiteTableCount(t, db, "user_data", 1)
	assertSQLiteTableCount(t, db, "user_measurements", 1)
	assertSQLiteTableCount(t, db, "sleep_records", 0)
	assertSQLiteTableCount(t, db, "recovery_records", 0)
	assertSQLiteTableCount(t, db, "workout_records", 0)
	assertSQLiteTableCount(t, db, "cycle_records", 0)
}

func openSQLiteFromBytes(t *testing.T, data []byte) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "output.sqlite")
	if err := os.WriteFile(dbPath, data, 0644); err != nil {
		t.Fatalf("failed to write sqlite bytes to file: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite database from bytes: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func assertSQLiteTableCount(t *testing.T, db *sql.DB, table string, expected int) {
	t.Helper()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatalf("failed to query table %s: %v", table, err)
	}

	if count != expected {
		t.Fatalf("unexpected row count for table %s. expected %d, got %d", table, expected, count)
	}
}
