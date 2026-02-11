// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

type sqliteFixture struct {
	userID                int
	email                 string
	firstName             string
	lastName              string
	sleepRecords          []string
	recoveryScore         float64
	userCalibrating       int
	workoutScoreState     string
	workoutAverageHR      int
	cycleAverageHeartRate int
	exportSource          string
}

func TestWriteSQLiteToFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "user.sqlite")
	data := createSQLiteFixtureData(t, sqliteFixture{
		userID:                111111111,
		email:                 "john.doe@example.com",
		firstName:             "John",
		lastName:              "Doe",
		sleepRecords:          []string{"sleep-1"},
		recoveryScore:         80,
		userCalibrating:       0,
		workoutScoreState:     "SCORED",
		workoutAverageHR:      120,
		cycleAverageHeartRate: 130,
		exportSource:          "run-1",
	})

	err := writeSQLiteToFile(filePath, data)
	if err != nil {
		t.Fatalf("writeSQLiteToFile returned an error: %v", err)
	}

	db := openSQLiteFile(t, filePath)
	assertSQLiteCount(t, db, "user_data", 1)
	assertSQLiteCount(t, db, "sleep_records", 1)
	assertSQLiteCount(t, db, "export_runs", 1)
}

func TestWriteSQLiteToFileIncrementalMerge(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "user.sqlite")

	initialData := createSQLiteFixtureData(t, sqliteFixture{
		userID:                111111111,
		email:                 "john.doe@example.com",
		firstName:             "John",
		lastName:              "Doe",
		sleepRecords:          []string{"sleep-1"},
		recoveryScore:         70,
		userCalibrating:       0,
		workoutScoreState:     "PENDING",
		workoutAverageHR:      110,
		cycleAverageHeartRate: 125,
		exportSource:          "run-1",
	})

	if err := writeSQLiteToFile(filePath, initialData); err != nil {
		t.Fatalf("writeSQLiteToFile initial call returned an error: %v", err)
	}

	incrementalData := createSQLiteFixtureData(t, sqliteFixture{
		userID:                111111111,
		email:                 "john.updated@example.com",
		firstName:             "John",
		lastName:              "Doe",
		sleepRecords:          []string{"sleep-1", "sleep-2"},
		recoveryScore:         85,
		userCalibrating:       1,
		workoutScoreState:     "SCORED",
		workoutAverageHR:      140,
		cycleAverageHeartRate: 155,
		exportSource:          "run-2",
	})

	if err := writeSQLiteToFile(filePath, incrementalData); err != nil {
		t.Fatalf("writeSQLiteToFile incremental call returned an error: %v", err)
	}

	db := openSQLiteFile(t, filePath)

	assertSQLiteCount(t, db, "sleep_records", 2)
	assertSQLiteCount(t, db, "export_runs", 2)

	var email string
	if err := db.QueryRow(`SELECT email FROM user_data WHERE user_id = ?`, 111111111).Scan(&email); err != nil {
		t.Fatalf("failed to query merged user_data: %v", err)
	}
	if email != "john.updated@example.com" {
		t.Fatalf("unexpected merged email value. expected john.updated@example.com, got %s", email)
	}

	var recoveryScore float64
	if err := db.QueryRow(`SELECT recovery_score FROM recovery_records WHERE cycle_id = ? AND sleep_id = ?`, 7001, "sleep-1").Scan(&recoveryScore); err != nil {
		t.Fatalf("failed to query merged recovery_records: %v", err)
	}
	if recoveryScore != 85 {
		t.Fatalf("unexpected merged recovery_score. expected 85, got %v", recoveryScore)
	}
}

func TestWriteSQLiteToFileEmptyData(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "user.sqlite")
	err := writeSQLiteToFile(filePath, []byte{})
	if err == nil {
		t.Fatal("expected error for empty sqlite payload, got nil")
	}
}

func TestWriteToFileSQLitePath(t *testing.T) {
	tempDir := t.TempDir()
	cfg := FileExport{
		FilePath:       tempDir,
		FileType:       "sqlite",
		FileName:       "user",
		FileNamePrefix: "",
		ServerMode:     false,
	}
	data := createSQLiteFixtureData(t, sqliteFixture{
		userID:                111111111,
		email:                 "john.doe@example.com",
		firstName:             "John",
		lastName:              "Doe",
		sleepRecords:          []string{"sleep-1"},
		recoveryScore:         80,
		userCalibrating:       0,
		workoutScoreState:     "SCORED",
		workoutAverageHR:      120,
		cycleAverageHeartRate: 130,
		exportSource:          "run-1",
	})

	err := writeToFile(cfg, data)
	if err != nil {
		t.Fatalf("writeToFile returned an error for sqlite file type: %v", err)
	}

	_, err = os.Stat(filepath.Join(tempDir, "user.sqlite"))
	if err != nil {
		t.Fatalf("expected sqlite file to exist, got error: %v", err)
	}
}

func createSQLiteFixtureData(t *testing.T, fixture sqliteFixture) []byte {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "fixture.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open fixture sqlite database: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to start fixture transaction: %v", err)
	}

	if err := createSQLiteMergeSchema(tx); err != nil {
		t.Fatalf("failed to create fixture sqlite schema: %v", err)
	}

	now := time.Date(2024, time.August, 15, 10, 30, 0, 0, time.UTC).Format(time.RFC3339Nano)

	if _, err := tx.Exec(`INSERT INTO user_data(user_id, email, first_name, last_name) VALUES (?, ?, ?, ?)`,
		fixture.userID,
		fixture.email,
		fixture.firstName,
		fixture.lastName,
	); err != nil {
		t.Fatalf("failed to insert fixture user_data: %v", err)
	}

	if _, err := tx.Exec(`INSERT INTO user_measurements(user_id, height_meter, weight_kilogram, max_heart_rate) VALUES (?, ?, ?, ?)`,
		fixture.userID,
		1.8,
		80.1,
		180,
	); err != nil {
		t.Fatalf("failed to insert fixture user_measurements: %v", err)
	}

	for index, sleepID := range fixture.sleepRecords {
		if _, err := tx.Exec(
			`INSERT INTO sleep_records(
				id, user_id, created_at, updated_at, start, end, timezone_offset, nap, score_state,
				total_in_bed_time_milli, total_awake_time_milli, total_no_data_time_milli,
				total_light_sleep_time_milli, total_slow_wave_sleep_time_milli, total_rem_sleep_time_milli,
				sleep_cycle_count, disturbance_count, baseline_milli, need_from_sleep_debt_milli,
				need_from_recent_strain_milli, need_from_recent_nap_milli, respiratory_rate,
				sleep_performance_percentage, sleep_consistency_percentage, sleep_efficiency_percentage
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			sleepID, fixture.userID, now, now, now, now, "-07:00", 0, "SCORED",
			1000, 1000, 1000, 1000, 1000, 1000, index+1, 0, 0, 0, 0, 0, 13.1, 80, 80, 80,
		); err != nil {
			t.Fatalf("failed to insert fixture sleep_records: %v", err)
		}
	}

	if _, err := tx.Exec(
		`INSERT INTO recovery_records(
			cycle_id, sleep_id, user_id, created_at, updated_at, score_state,
			user_calibrating, recovery_score, resting_heart_rate, hrv_rmssd_milli,
			spo2_percentage, skin_temp_celsius
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		7001, "sleep-1", fixture.userID, now, now, "SCORED", fixture.userCalibrating, fixture.recoveryScore, 50.0, 90.0, 98.0, 36.9,
	); err != nil {
		t.Fatalf("failed to insert fixture recovery_records: %v", err)
	}

	if _, err := tx.Exec(
		`INSERT INTO workout_records(
			id, user_id, created_at, updated_at, start, end, timezone_offset, sport_id, sport_name, score_state,
			strain, average_heart_rate, max_heart_rate, kilojoule, percent_recorded, distance_meter,
			altitude_gain_meter, altitude_change_meter, zone_zero_milli, zone_one_milli, zone_two_milli,
			zone_three_milli, zone_four_milli, zone_five_milli
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"workout-1", fixture.userID, now, now, now, now, "-07:00", 44, "Running", fixture.workoutScoreState,
		7.1, fixture.workoutAverageHR, 180, 500.0, 100, 10000, 200, 100, 1, 2, 3, 4, 5, 6,
	); err != nil {
		t.Fatalf("failed to insert fixture workout_records: %v", err)
	}

	if _, err := tx.Exec(
		`INSERT INTO cycle_records(
			id, user_id, created_at, updated_at, start, end, timezone_offset, score_state,
			strain, kilojoule, average_heart_rate, max_heart_rate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		9001, fixture.userID, now, now, now, now, "-07:00", "SCORED", 8.0, 600.0, fixture.cycleAverageHeartRate, 170,
	); err != nil {
		t.Fatalf("failed to insert fixture cycle_records: %v", err)
	}

	if _, err := tx.Exec(`INSERT INTO export_runs(exported_at, source) VALUES (?, ?)`, now, fixture.exportSource); err != nil {
		t.Fatalf("failed to insert fixture export_runs: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("failed to commit fixture transaction: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("failed to close fixture sqlite db: %v", err)
	}

	data, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("failed to read fixture sqlite file: %v", err)
	}

	return data
}

func openSQLiteFile(t *testing.T, filePath string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		t.Fatalf("failed to open sqlite file: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func assertSQLiteCount(t *testing.T, db *sql.DB, table string, expected int) {
	t.Helper()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatalf("failed to query table %s: %v", table, err)
	}

	if count != expected {
		t.Fatalf("unexpected row count for table %s. expected %d, got %d", table, expected, count)
	}
}
