// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const sqliteBusyTimeoutMS = 5000

var sqliteDataTableNames = []string{
	"user_data",
	"user_measurements",
	"sleep_records",
	"recovery_records",
	"workout_records",
	"cycle_records",
}

var sqliteSchemaStatementsByTable = map[string]string{
	"user_data": `CREATE TABLE IF NOT EXISTS user_data (
			user_id INTEGER PRIMARY KEY,
			email TEXT NOT NULL,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL
		)`,
	"user_measurements": `CREATE TABLE IF NOT EXISTS user_measurements (
			user_id INTEGER PRIMARY KEY,
			height_meter REAL NOT NULL,
			weight_kilogram REAL NOT NULL,
			max_heart_rate INTEGER NOT NULL
		)`,
	"sleep_records": `CREATE TABLE IF NOT EXISTS sleep_records (
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
	"recovery_records": `CREATE TABLE IF NOT EXISTS recovery_records (
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
	"workout_records": `CREATE TABLE IF NOT EXISTS workout_records (
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
	"cycle_records": `CREATE TABLE IF NOT EXISTS cycle_records (
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
}

const sqliteExportRunSchema = `CREATE TABLE IF NOT EXISTS export_runs (
			run_id INTEGER PRIMARY KEY AUTOINCREMENT,
			exported_at TEXT NOT NULL,
			source TEXT NOT NULL
		)`

// ConvertToSQLite converts user data to a SQLite database file and returns the raw bytes.
func ConvertToSQLite(userData User) ([]byte, error) {
	tempFile, err := os.CreateTemp("", "mywhoop-*.sqlite")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary sqlite file: %w", err)
	}

	tempFilePath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary sqlite file: %w", err)
	}

	defer func() {
		_ = os.Remove(tempFilePath)
	}()

	db, err := sql.Open("sqlite", tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := configureSQLiteConnection(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := writeUserDataToSQLite(db, userData); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := db.Close(); err != nil {
		return nil, fmt.Errorf("failed to close sqlite database: %w", err)
	}

	sqliteData, err := os.ReadFile(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sqlite file: %w", err)
	}

	return sqliteData, nil
}

func configureSQLiteConnection(db *sql.DB) error {
	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMS)); err != nil {
		return fmt.Errorf("failed to configure sqlite busy timeout: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable sqlite foreign keys: %w", err)
	}

	return nil
}

func writeUserDataToSQLite(db *sql.DB, userData User) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start sqlite transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := createSQLiteSchema(tx); err != nil {
		return err
	}

	if err := upsertSQLiteUserData(tx, userData); err != nil {
		return err
	}

	if err := upsertSQLiteUserMeasurements(tx, userData); err != nil {
		return err
	}

	if err := upsertSQLiteSleepRecords(tx, userData.SleepCollection.SleepCollectionRecords); err != nil {
		return err
	}

	if err := upsertSQLiteRecoveryRecords(tx, userData.RecoveryCollection.RecoveryRecords); err != nil {
		return err
	}

	if err := upsertSQLiteWorkoutRecords(tx, userData.WorkoutCollection.Records); err != nil {
		return err
	}

	if err := upsertSQLiteCycleRecords(tx, userData.CycleCollection.Records); err != nil {
		return err
	}

	if _, err := tx.Exec(`INSERT INTO export_runs(exported_at, source) VALUES (?, ?)`, formatSQLiteTimestamp(time.Now().UTC()), "mywhoop"); err != nil {
		return fmt.Errorf("failed to insert sqlite export run metadata: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit sqlite transaction: %w", err)
	}

	return nil
}

func createSQLiteSchema(tx *sql.Tx) error {
	for _, table := range sqliteDataTableNames {
		stmt, ok := sqliteSchemaStatementsByTable[table]
		if !ok {
			return fmt.Errorf("sqlite schema statement missing for table %s", table)
		}

		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to initialize sqlite schema for table %s: %w", table, err)
		}
	}

	if _, err := tx.Exec(sqliteExportRunSchema); err != nil {
		return fmt.Errorf("failed to initialize sqlite export metadata schema: %w", err)
	}

	return nil
}

func upsertSQLiteUserData(tx *sql.Tx, userData User) error {
	_, err := tx.Exec(`
		INSERT INTO user_data(user_id, email, first_name, last_name)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			email = excluded.email,
			first_name = excluded.first_name,
			last_name = excluded.last_name`,
		userData.UserData.UserID,
		userData.UserData.Email,
		userData.UserData.FirstName,
		userData.UserData.LastName,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert user data: %w", err)
	}

	return nil
}

func upsertSQLiteUserMeasurements(tx *sql.Tx, userData User) error {
	_, err := tx.Exec(`
		INSERT INTO user_measurements(user_id, height_meter, weight_kilogram, max_heart_rate)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			height_meter = excluded.height_meter,
			weight_kilogram = excluded.weight_kilogram,
			max_heart_rate = excluded.max_heart_rate`,
		userData.UserData.UserID,
		userData.UserMeasurements.HeightMeter,
		userData.UserMeasurements.WeightKilogram,
		userData.UserMeasurements.MaxHeartRate,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert user measurements: %w", err)
	}

	return nil
}

func upsertSQLiteSleepRecords(tx *sql.Tx, records []SleepCollectionRecords) error {
	insertStmt := `
		INSERT INTO sleep_records(
			id, user_id, created_at, updated_at, start, end, timezone_offset, nap, score_state,
			total_in_bed_time_milli, total_awake_time_milli, total_no_data_time_milli,
			total_light_sleep_time_milli, total_slow_wave_sleep_time_milli, total_rem_sleep_time_milli,
			sleep_cycle_count, disturbance_count, baseline_milli, need_from_sleep_debt_milli,
			need_from_recent_strain_milli, need_from_recent_nap_milli, respiratory_rate,
			sleep_performance_percentage, sleep_consistency_percentage, sleep_efficiency_percentage
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id = excluded.user_id,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			start = excluded.start,
			end = excluded.end,
			timezone_offset = excluded.timezone_offset,
			nap = excluded.nap,
			score_state = excluded.score_state,
			total_in_bed_time_milli = excluded.total_in_bed_time_milli,
			total_awake_time_milli = excluded.total_awake_time_milli,
			total_no_data_time_milli = excluded.total_no_data_time_milli,
			total_light_sleep_time_milli = excluded.total_light_sleep_time_milli,
			total_slow_wave_sleep_time_milli = excluded.total_slow_wave_sleep_time_milli,
			total_rem_sleep_time_milli = excluded.total_rem_sleep_time_milli,
			sleep_cycle_count = excluded.sleep_cycle_count,
			disturbance_count = excluded.disturbance_count,
			baseline_milli = excluded.baseline_milli,
			need_from_sleep_debt_milli = excluded.need_from_sleep_debt_milli,
			need_from_recent_strain_milli = excluded.need_from_recent_strain_milli,
			need_from_recent_nap_milli = excluded.need_from_recent_nap_milli,
			respiratory_rate = excluded.respiratory_rate,
			sleep_performance_percentage = excluded.sleep_performance_percentage,
			sleep_consistency_percentage = excluded.sleep_consistency_percentage,
			sleep_efficiency_percentage = excluded.sleep_efficiency_percentage`

	for _, record := range records {
		if _, err := tx.Exec(
			insertStmt,
			record.ID,
			record.UserID,
			formatSQLiteTimestamp(record.CreatedAt),
			formatSQLiteTimestamp(record.UpdatedAt),
			formatSQLiteTimestamp(record.Start),
			formatSQLiteTimestamp(record.End),
			record.TimezoneOffset,
			boolToSQLiteInt(record.Nap),
			record.ScoreState,
			record.Score.StageSummary.TotalInBedTimeMilli,
			record.Score.StageSummary.TotalAwakeTimeMilli,
			record.Score.StageSummary.TotalNoDataTimeMilli,
			record.Score.StageSummary.TotalLightSleepTimeMilli,
			record.Score.StageSummary.TotalSlowWaveSleepTimeMilli,
			record.Score.StageSummary.TotalRemSleepTimeMilli,
			record.Score.StageSummary.SleepCycleCount,
			record.Score.StageSummary.DisturbanceCount,
			record.Score.SleepNeeded.BaselineMilli,
			record.Score.SleepNeeded.NeedFromSleepDebtMilli,
			record.Score.SleepNeeded.NeedFromRecentStrainMilli,
			record.Score.SleepNeeded.NeedFromRecentNapMilli,
			record.Score.RespiratoryRate,
			record.Score.SleepPerformancePercentage,
			record.Score.SleepConsistencyPercentage,
			record.Score.SleepEfficiencyPercentage,
		); err != nil {
			return fmt.Errorf("failed to upsert sleep record %s: %w", record.ID, err)
		}
	}

	return nil
}

func upsertSQLiteRecoveryRecords(tx *sql.Tx, records []RecoveryRecords) error {
	insertStmt := `
		INSERT INTO recovery_records(
			cycle_id, sleep_id, user_id, created_at, updated_at, score_state,
			user_calibrating, recovery_score, resting_heart_rate, hrv_rmssd_milli,
			spo2_percentage, skin_temp_celsius
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(cycle_id, sleep_id) DO UPDATE SET
			user_id = excluded.user_id,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			score_state = excluded.score_state,
			user_calibrating = excluded.user_calibrating,
			recovery_score = excluded.recovery_score,
			resting_heart_rate = excluded.resting_heart_rate,
			hrv_rmssd_milli = excluded.hrv_rmssd_milli,
			spo2_percentage = excluded.spo2_percentage,
			skin_temp_celsius = excluded.skin_temp_celsius`

	for _, record := range records {
		if _, err := tx.Exec(
			insertStmt,
			record.CycleID,
			record.SleepID,
			record.UserID,
			formatSQLiteTimestamp(record.CreatedAt),
			formatSQLiteTimestamp(record.UpdatedAt),
			record.ScoreState,
			boolToSQLiteInt(record.Score.UserCalibrating),
			record.Score.RecoveryScore,
			record.Score.RestingHeartRate,
			record.Score.HrvRmssdMilli,
			record.Score.Spo2Percentage,
			record.Score.SkinTempCelsius,
		); err != nil {
			return fmt.Errorf("failed to upsert recovery record cycle_id=%d sleep_id=%s: %w", record.CycleID, record.SleepID, err)
		}
	}

	return nil
}

func upsertSQLiteWorkoutRecords(tx *sql.Tx, records []WorkoutRecords) error {
	insertStmt := `
		INSERT INTO workout_records(
			id, user_id, created_at, updated_at, start, end, timezone_offset, sport_id, sport_name, score_state,
			strain, average_heart_rate, max_heart_rate, kilojoule, percent_recorded, distance_meter,
			altitude_gain_meter, altitude_change_meter, zone_zero_milli, zone_one_milli, zone_two_milli,
			zone_three_milli, zone_four_milli, zone_five_milli
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id = excluded.user_id,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			start = excluded.start,
			end = excluded.end,
			timezone_offset = excluded.timezone_offset,
			sport_id = excluded.sport_id,
			sport_name = excluded.sport_name,
			score_state = excluded.score_state,
			strain = excluded.strain,
			average_heart_rate = excluded.average_heart_rate,
			max_heart_rate = excluded.max_heart_rate,
			kilojoule = excluded.kilojoule,
			percent_recorded = excluded.percent_recorded,
			distance_meter = excluded.distance_meter,
			altitude_gain_meter = excluded.altitude_gain_meter,
			altitude_change_meter = excluded.altitude_change_meter,
			zone_zero_milli = excluded.zone_zero_milli,
			zone_one_milli = excluded.zone_one_milli,
			zone_two_milli = excluded.zone_two_milli,
			zone_three_milli = excluded.zone_three_milli,
			zone_four_milli = excluded.zone_four_milli,
			zone_five_milli = excluded.zone_five_milli`

	for _, record := range records {
		if _, err := tx.Exec(
			insertStmt,
			record.ID,
			record.UserID,
			formatSQLiteTimestamp(record.CreatedAt),
			formatSQLiteTimestamp(record.UpdatedAt),
			formatSQLiteTimestamp(record.Start),
			formatSQLiteTimestamp(record.End),
			record.TimezoneOffset,
			record.SportID,
			record.SportName,
			record.ScoreState,
			record.Score.Strain,
			record.Score.AverageHeartRate,
			record.Score.MaxHeartRate,
			record.Score.Kilojoule,
			record.Score.PercentRecorded,
			record.Score.DistanceMeter,
			record.Score.AltitudeGainMeter,
			record.Score.AltitudeChangeMeter,
			record.Score.ZoneDuration.ZoneZeroMilli,
			record.Score.ZoneDuration.ZoneOneMilli,
			record.Score.ZoneDuration.ZoneTwoMilli,
			record.Score.ZoneDuration.ZoneThreeMilli,
			record.Score.ZoneDuration.ZoneFourMilli,
			record.Score.ZoneDuration.ZoneFiveMilli,
		); err != nil {
			return fmt.Errorf("failed to upsert workout record %s: %w", record.ID, err)
		}
	}

	return nil
}

func upsertSQLiteCycleRecords(tx *sql.Tx, records []CycleRecords) error {
	insertStmt := `
		INSERT INTO cycle_records(
			id, user_id, created_at, updated_at, start, end, timezone_offset, score_state,
			strain, kilojoule, average_heart_rate, max_heart_rate
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id = excluded.user_id,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			start = excluded.start,
			end = excluded.end,
			timezone_offset = excluded.timezone_offset,
			score_state = excluded.score_state,
			strain = excluded.strain,
			kilojoule = excluded.kilojoule,
			average_heart_rate = excluded.average_heart_rate,
			max_heart_rate = excluded.max_heart_rate`

	for _, record := range records {
		if _, err := tx.Exec(
			insertStmt,
			record.ID,
			record.UserID,
			formatSQLiteTimestamp(record.CreatedAt),
			formatSQLiteTimestamp(record.UpdatedAt),
			formatSQLiteTimestamp(record.Start),
			formatSQLiteTimestamp(record.End),
			record.TimezoneOffset,
			record.ScoreState,
			record.Score.Strain,
			record.Score.Kilojoule,
			record.Score.AverageHeartRate,
			record.Score.MaxHeartRate,
		); err != nil {
			return fmt.Errorf("failed to upsert cycle record %d: %w", record.ID, err)
		}
	}

	return nil
}

func formatSQLiteTimestamp(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func boolToSQLiteInt(v bool) int {
	if v {
		return 1
	}

	return 0
}
