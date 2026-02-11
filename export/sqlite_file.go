// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

const sqliteBusyTimeoutMS = 5000

func writeSQLiteToFile(filePath string, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("sqlite export data is empty")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("unable to write sqlite file: %w", err)
		}

		return nil
	}

	return mergeSQLiteFile(filePath, data)
}

func mergeSQLiteFile(existingFilePath string, incomingData []byte) error {
	tempIncomingFile, err := os.CreateTemp("", "mywhoop-merge-*.sqlite")
	if err != nil {
		return fmt.Errorf("unable to create temporary sqlite file: %w", err)
	}

	tempIncomingFilePath := tempIncomingFile.Name()
	if err := tempIncomingFile.Close(); err != nil {
		return fmt.Errorf("unable to close temporary sqlite file: %w", err)
	}

	defer func() {
		_ = os.Remove(tempIncomingFilePath)
	}()

	if err := os.WriteFile(tempIncomingFilePath, incomingData, 0600); err != nil {
		return fmt.Errorf("unable to write temporary sqlite file: %w", err)
	}

	db, err := sql.Open("sqlite", existingFilePath)
	if err != nil {
		return fmt.Errorf("unable to open target sqlite file: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := configureSQLiteConnection(db); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("unable to start sqlite merge transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := createSQLiteMergeSchema(tx); err != nil {
		return err
	}

	escapedPath := strings.ReplaceAll(tempIncomingFilePath, "'", "''")
	if _, err := tx.Exec("ATTACH DATABASE '" + escapedPath + "' AS incoming"); err != nil {
		return fmt.Errorf("unable to attach incoming sqlite file: %w", err)
	}

	for _, table := range []string{
		"user_data",
		"user_measurements",
		"sleep_records",
		"recovery_records",
		"workout_records",
		"cycle_records",
	} {
		query := fmt.Sprintf("INSERT OR REPLACE INTO %s SELECT * FROM incoming.%s", table, table)
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("unable to merge sqlite table %s: %w", table, err)
		}
	}

	if _, err := tx.Exec("INSERT INTO export_runs(exported_at, source) SELECT exported_at, source FROM incoming.export_runs"); err != nil {
		return fmt.Errorf("unable to merge sqlite export metadata: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("unable to commit sqlite merge transaction: %w", err)
	}

	return nil
}

func configureSQLiteConnection(db *sql.DB) error {
	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMS)); err != nil {
		return fmt.Errorf("unable to configure sqlite busy timeout: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("unable to enable sqlite foreign keys: %w", err)
	}

	return nil
}

func createSQLiteMergeSchema(tx *sql.Tx) error {
	schemaStatements := []string{
		`CREATE TABLE IF NOT EXISTS user_data (
			user_id INTEGER PRIMARY KEY,
			email TEXT NOT NULL,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS user_measurements (
			user_id INTEGER PRIMARY KEY,
			height_meter REAL NOT NULL,
			weight_kilogram REAL NOT NULL,
			max_heart_rate INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sleep_records (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			start TEXT NOT NULL,
			end TEXT NOT NULL,
			timezone_offset TEXT NOT NULL,
			nap INTEGER NOT NULL,
			score_state TEXT NOT NULL,
			total_in_bed_time_milli INTEGER NOT NULL,
			total_awake_time_milli INTEGER NOT NULL,
			total_no_data_time_milli INTEGER NOT NULL,
			total_light_sleep_time_milli INTEGER NOT NULL,
			total_slow_wave_sleep_time_milli INTEGER NOT NULL,
			total_rem_sleep_time_milli INTEGER NOT NULL,
			sleep_cycle_count INTEGER NOT NULL,
			disturbance_count INTEGER NOT NULL,
			baseline_milli INTEGER NOT NULL,
			need_from_sleep_debt_milli INTEGER NOT NULL,
			need_from_recent_strain_milli INTEGER NOT NULL,
			need_from_recent_nap_milli INTEGER NOT NULL,
			respiratory_rate REAL NOT NULL,
			sleep_performance_percentage REAL NOT NULL,
			sleep_consistency_percentage REAL NOT NULL,
			sleep_efficiency_percentage REAL NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS recovery_records (
			cycle_id INTEGER NOT NULL,
			sleep_id TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			score_state TEXT NOT NULL,
			user_calibrating INTEGER NOT NULL,
			recovery_score REAL NOT NULL,
			resting_heart_rate REAL NOT NULL,
			hrv_rmssd_milli REAL NOT NULL,
			spo2_percentage REAL NOT NULL,
			skin_temp_celsius REAL NOT NULL,
			PRIMARY KEY (cycle_id, sleep_id)
		)`,
		`CREATE TABLE IF NOT EXISTS workout_records (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			start TEXT NOT NULL,
			end TEXT NOT NULL,
			timezone_offset TEXT NOT NULL,
			sport_id INTEGER NOT NULL,
			sport_name TEXT NOT NULL,
			score_state TEXT NOT NULL,
			strain REAL NOT NULL,
			average_heart_rate INTEGER NOT NULL,
			max_heart_rate INTEGER NOT NULL,
			kilojoule REAL NOT NULL,
			percent_recorded REAL NOT NULL,
			distance_meter REAL NOT NULL,
			altitude_gain_meter REAL NOT NULL,
			altitude_change_meter REAL NOT NULL,
			zone_zero_milli INTEGER NOT NULL,
			zone_one_milli INTEGER NOT NULL,
			zone_two_milli INTEGER NOT NULL,
			zone_three_milli INTEGER NOT NULL,
			zone_four_milli INTEGER NOT NULL,
			zone_five_milli INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS cycle_records (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			start TEXT NOT NULL,
			end TEXT NOT NULL,
			timezone_offset TEXT NOT NULL,
			score_state TEXT NOT NULL,
			strain REAL NOT NULL,
			kilojoule REAL NOT NULL,
			average_heart_rate INTEGER NOT NULL,
			max_heart_rate INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS export_runs (
			run_id INTEGER PRIMARY KEY AUTOINCREMENT,
			exported_at TEXT NOT NULL,
			source TEXT NOT NULL
		)`,
	}

	for _, stmt := range schemaStatements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("unable to initialize sqlite schema for file merge: %w", err)
		}
	}

	return nil
}
