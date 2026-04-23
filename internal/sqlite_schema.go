// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"database/sql"
	"fmt"
)

// schemaStatements is the source of truth for the sqlite schema.
// Tables are created in dependency order and indexes are added after.
//
// Design notes:
//   - Timestamps are stored as RFC3339Nano text (portable, human-readable, sortable).
//   - Booleans are stored as INTEGER (0/1).
//   - Nullable columns use NULL; everything else is NOT NULL.
//   - No foreign key constraints declared today; PRAGMA foreign_keys is on so they can be
//     added later without a schema rewrite. Recovery records can arrive in pagination
//     windows that don't include their referenced sleep/cycle, so hard FKs would cause
//     avoidable insertion failures.
var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS user_data (
		user_id    INTEGER PRIMARY KEY,
		email      TEXT NOT NULL,
		first_name TEXT NOT NULL,
		last_name  TEXT NOT NULL
	)`,

	`CREATE TABLE IF NOT EXISTS user_measurements (
		user_id         INTEGER PRIMARY KEY,
		height_meter    REAL NOT NULL,
		weight_kilogram REAL NOT NULL,
		max_heart_rate  INTEGER NOT NULL
	)`,

	`CREATE TABLE IF NOT EXISTS sleep_records (
		id                               TEXT    PRIMARY KEY,
		user_id                          INTEGER NOT NULL,
		created_at                       TEXT    NOT NULL,
		updated_at                       TEXT    NOT NULL,
		start                            TEXT    NOT NULL,
		end                              TEXT    NOT NULL,
		timezone_offset                  TEXT    NOT NULL,
		nap                              INTEGER NOT NULL,
		score_state                      TEXT    NOT NULL,
		total_in_bed_time_milli          INTEGER NOT NULL,
		total_awake_time_milli           INTEGER NOT NULL,
		total_no_data_time_milli         INTEGER NOT NULL,
		total_light_sleep_time_milli     INTEGER NOT NULL,
		total_slow_wave_sleep_time_milli INTEGER NOT NULL,
		total_rem_sleep_time_milli       INTEGER NOT NULL,
		sleep_cycle_count                INTEGER NOT NULL,
		disturbance_count                INTEGER NOT NULL,
		baseline_milli                   INTEGER NOT NULL,
		need_from_sleep_debt_milli       INTEGER NOT NULL,
		need_from_recent_strain_milli    INTEGER NOT NULL,
		need_from_recent_nap_milli       INTEGER NOT NULL,
		respiratory_rate                 REAL    NOT NULL,
		sleep_performance_percentage     REAL    NOT NULL,
		sleep_consistency_percentage     REAL    NOT NULL,
		sleep_efficiency_percentage      REAL    NOT NULL
	)`,

	`CREATE TABLE IF NOT EXISTS cycle_records (
		id                 INTEGER PRIMARY KEY,
		user_id            INTEGER NOT NULL,
		created_at         TEXT    NOT NULL,
		updated_at         TEXT    NOT NULL,
		start              TEXT    NOT NULL,
		end                TEXT    NOT NULL,
		timezone_offset    TEXT    NOT NULL,
		score_state        TEXT    NOT NULL,
		strain             REAL    NOT NULL,
		kilojoule          REAL    NOT NULL,
		average_heart_rate INTEGER NOT NULL,
		max_heart_rate     INTEGER NOT NULL
	)`,

	`CREATE TABLE IF NOT EXISTS recovery_records (
		cycle_id           INTEGER NOT NULL,
		sleep_id           TEXT    NOT NULL,
		user_id            INTEGER NOT NULL,
		created_at         TEXT    NOT NULL,
		updated_at         TEXT    NOT NULL,
		score_state        TEXT    NOT NULL,
		user_calibrating   INTEGER NOT NULL,
		recovery_score     REAL    NOT NULL,
		resting_heart_rate REAL    NOT NULL,
		hrv_rmssd_milli    REAL    NOT NULL,
		spo2_percentage    REAL    NOT NULL,
		skin_temp_celsius  REAL    NOT NULL,
		PRIMARY KEY (cycle_id, sleep_id)
	)`,

	`CREATE TABLE IF NOT EXISTS workout_records (
		id                    TEXT    PRIMARY KEY,
		user_id               INTEGER NOT NULL,
		created_at            TEXT    NOT NULL,
		updated_at            TEXT    NOT NULL,
		start                 TEXT    NOT NULL,
		end                   TEXT    NOT NULL,
		timezone_offset       TEXT    NOT NULL,
		sport_id              INTEGER,
		sport_name            TEXT,
		score_state           TEXT    NOT NULL,
		strain                REAL    NOT NULL,
		average_heart_rate    INTEGER NOT NULL,
		max_heart_rate        INTEGER NOT NULL,
		kilojoule             REAL    NOT NULL,
		percent_recorded      REAL    NOT NULL,
		distance_meter        REAL    NOT NULL,
		altitude_gain_meter   REAL    NOT NULL,
		altitude_change_meter REAL    NOT NULL,
		zone_zero_milli       INTEGER NOT NULL,
		zone_one_milli        INTEGER NOT NULL,
		zone_two_milli        INTEGER NOT NULL,
		zone_three_milli      INTEGER NOT NULL,
		zone_four_milli       INTEGER NOT NULL,
		zone_five_milli       INTEGER NOT NULL
	)`,

	`CREATE INDEX IF NOT EXISTS idx_sleep_records_user_start    ON sleep_records(user_id, start)`,
	`CREATE INDEX IF NOT EXISTS idx_cycle_records_user_start    ON cycle_records(user_id, start)`,
	`CREATE INDEX IF NOT EXISTS idx_workout_records_user_start  ON workout_records(user_id, start)`,
	`CREATE INDEX IF NOT EXISTS idx_recovery_records_user_sleep ON recovery_records(user_id, sleep_id)`,
}

func applySQLiteSchema(db *sql.DB) error {
	for _, stmt := range schemaStatements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("apply sqlite schema: %w", err)
		}
	}
	return nil
}

// upsertUser writes every piece of the User payload into its table.
func upsertUser(tx *sql.Tx, u User) error {
	if u.UserData.UserID != 0 {
		if err := upsertUserData(tx, u.UserData); err != nil {
			return err
		}
	}

	if err := upsertUserMeasurements(tx, u.UserData.UserID, u.UserMeasurements); err != nil {
		return err
	}

	if err := upsertSleepRecords(tx, u.SleepCollection.SleepCollectionRecords); err != nil {
		return err
	}

	if err := upsertCycleRecords(tx, u.CycleCollection.Records); err != nil {
		return err
	}

	if err := upsertRecoveryRecords(tx, u.RecoveryCollection.RecoveryRecords); err != nil {
		return err
	}

	if err := upsertWorkoutRecords(tx, u.WorkoutCollection.Records); err != nil {
		return err
	}

	return nil
}

func upsertUserData(tx *sql.Tx, ud UserData) error {
	_, err := tx.Exec(`
		INSERT INTO user_data(user_id, email, first_name, last_name)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			email      = excluded.email,
			first_name = excluded.first_name,
			last_name  = excluded.last_name`,
		ud.UserID, ud.Email, ud.FirstName, ud.LastName,
	)
	if err != nil {
		return fmt.Errorf("upsert user_data: %w", err)
	}
	return nil
}

func upsertUserMeasurements(tx *sql.Tx, userID int, m UserMeasurements) error {
	if userID == 0 {
		return nil
	}
	_, err := tx.Exec(`
		INSERT INTO user_measurements(user_id, height_meter, weight_kilogram, max_heart_rate)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			height_meter    = excluded.height_meter,
			weight_kilogram = excluded.weight_kilogram,
			max_heart_rate  = excluded.max_heart_rate`,
		userID, m.HeightMeter, m.WeightKilogram, m.MaxHeartRate,
	)
	if err != nil {
		return fmt.Errorf("upsert user_measurements: %w", err)
	}
	return nil
}

func upsertSleepRecords(tx *sql.Tx, records []SleepCollectionRecords) error {
	const stmt = `
		INSERT INTO sleep_records(
			id, user_id, created_at, updated_at, start, end, timezone_offset, nap, score_state,
			total_in_bed_time_milli, total_awake_time_milli, total_no_data_time_milli,
			total_light_sleep_time_milli, total_slow_wave_sleep_time_milli, total_rem_sleep_time_milli,
			sleep_cycle_count, disturbance_count, baseline_milli, need_from_sleep_debt_milli,
			need_from_recent_strain_milli, need_from_recent_nap_milli, respiratory_rate,
			sleep_performance_percentage, sleep_consistency_percentage, sleep_efficiency_percentage
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id                          = excluded.user_id,
			created_at                       = excluded.created_at,
			updated_at                       = excluded.updated_at,
			start                            = excluded.start,
			end                              = excluded.end,
			timezone_offset                  = excluded.timezone_offset,
			nap                              = excluded.nap,
			score_state                      = excluded.score_state,
			total_in_bed_time_milli          = excluded.total_in_bed_time_milli,
			total_awake_time_milli           = excluded.total_awake_time_milli,
			total_no_data_time_milli         = excluded.total_no_data_time_milli,
			total_light_sleep_time_milli     = excluded.total_light_sleep_time_milli,
			total_slow_wave_sleep_time_milli = excluded.total_slow_wave_sleep_time_milli,
			total_rem_sleep_time_milli       = excluded.total_rem_sleep_time_milli,
			sleep_cycle_count                = excluded.sleep_cycle_count,
			disturbance_count                = excluded.disturbance_count,
			baseline_milli                   = excluded.baseline_milli,
			need_from_sleep_debt_milli       = excluded.need_from_sleep_debt_milli,
			need_from_recent_strain_milli    = excluded.need_from_recent_strain_milli,
			need_from_recent_nap_milli       = excluded.need_from_recent_nap_milli,
			respiratory_rate                 = excluded.respiratory_rate,
			sleep_performance_percentage     = excluded.sleep_performance_percentage,
			sleep_consistency_percentage     = excluded.sleep_consistency_percentage,
			sleep_efficiency_percentage      = excluded.sleep_efficiency_percentage`

	for _, r := range records {
		if _, err := tx.Exec(stmt,
			r.ID, r.UserID, formatSQLiteTime(r.CreatedAt), formatSQLiteTime(r.UpdatedAt),
			formatSQLiteTime(r.Start), formatSQLiteTime(r.End), r.TimezoneOffset,
			boolToSQLiteInt(r.Nap), r.ScoreState,
			r.Score.StageSummary.TotalInBedTimeMilli, r.Score.StageSummary.TotalAwakeTimeMilli,
			r.Score.StageSummary.TotalNoDataTimeMilli, r.Score.StageSummary.TotalLightSleepTimeMilli,
			r.Score.StageSummary.TotalSlowWaveSleepTimeMilli, r.Score.StageSummary.TotalRemSleepTimeMilli,
			r.Score.StageSummary.SleepCycleCount, r.Score.StageSummary.DisturbanceCount,
			r.Score.SleepNeeded.BaselineMilli, r.Score.SleepNeeded.NeedFromSleepDebtMilli,
			r.Score.SleepNeeded.NeedFromRecentStrainMilli, r.Score.SleepNeeded.NeedFromRecentNapMilli,
			r.Score.RespiratoryRate, r.Score.SleepPerformancePercentage,
			r.Score.SleepConsistencyPercentage, r.Score.SleepEfficiencyPercentage,
		); err != nil {
			return fmt.Errorf("upsert sleep_records id=%s: %w", r.ID, err)
		}
	}
	return nil
}

func upsertCycleRecords(tx *sql.Tx, records []CycleRecords) error {
	const stmt = `
		INSERT INTO cycle_records(
			id, user_id, created_at, updated_at, start, end, timezone_offset, score_state,
			strain, kilojoule, average_heart_rate, max_heart_rate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id            = excluded.user_id,
			created_at         = excluded.created_at,
			updated_at         = excluded.updated_at,
			start              = excluded.start,
			end                = excluded.end,
			timezone_offset    = excluded.timezone_offset,
			score_state        = excluded.score_state,
			strain             = excluded.strain,
			kilojoule          = excluded.kilojoule,
			average_heart_rate = excluded.average_heart_rate,
			max_heart_rate     = excluded.max_heart_rate`

	for _, r := range records {
		if _, err := tx.Exec(stmt,
			r.ID, r.UserID, formatSQLiteTime(r.CreatedAt), formatSQLiteTime(r.UpdatedAt),
			formatSQLiteTime(r.Start), formatSQLiteTime(r.End), r.TimezoneOffset, r.ScoreState,
			r.Score.Strain, r.Score.Kilojoule, r.Score.AverageHeartRate, r.Score.MaxHeartRate,
		); err != nil {
			return fmt.Errorf("upsert cycle_records id=%d: %w", r.ID, err)
		}
	}
	return nil
}

func upsertRecoveryRecords(tx *sql.Tx, records []RecoveryRecords) error {
	const stmt = `
		INSERT INTO recovery_records(
			cycle_id, sleep_id, user_id, created_at, updated_at, score_state,
			user_calibrating, recovery_score, resting_heart_rate, hrv_rmssd_milli,
			spo2_percentage, skin_temp_celsius
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(cycle_id, sleep_id) DO UPDATE SET
			user_id            = excluded.user_id,
			created_at         = excluded.created_at,
			updated_at         = excluded.updated_at,
			score_state        = excluded.score_state,
			user_calibrating   = excluded.user_calibrating,
			recovery_score     = excluded.recovery_score,
			resting_heart_rate = excluded.resting_heart_rate,
			hrv_rmssd_milli    = excluded.hrv_rmssd_milli,
			spo2_percentage    = excluded.spo2_percentage,
			skin_temp_celsius  = excluded.skin_temp_celsius`

	for _, r := range records {
		if _, err := tx.Exec(stmt,
			r.CycleID, r.SleepID, r.UserID,
			formatSQLiteTime(r.CreatedAt), formatSQLiteTime(r.UpdatedAt),
			r.ScoreState, boolToSQLiteInt(r.Score.UserCalibrating),
			r.Score.RecoveryScore, r.Score.RestingHeartRate, r.Score.HrvRmssdMilli,
			r.Score.Spo2Percentage, r.Score.SkinTempCelsius,
		); err != nil {
			return fmt.Errorf("upsert recovery_records cycle_id=%d sleep_id=%s: %w", r.CycleID, r.SleepID, err)
		}
	}
	return nil
}

func upsertWorkoutRecords(tx *sql.Tx, records []WorkoutRecords) error {
	const stmt = `
		INSERT INTO workout_records(
			id, user_id, created_at, updated_at, start, end, timezone_offset,
			sport_id, sport_name, score_state,
			strain, average_heart_rate, max_heart_rate, kilojoule, percent_recorded,
			distance_meter, altitude_gain_meter, altitude_change_meter,
			zone_zero_milli, zone_one_milli, zone_two_milli,
			zone_three_milli, zone_four_milli, zone_five_milli
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id               = excluded.user_id,
			created_at            = excluded.created_at,
			updated_at            = excluded.updated_at,
			start                 = excluded.start,
			end                   = excluded.end,
			timezone_offset       = excluded.timezone_offset,
			sport_id              = excluded.sport_id,
			sport_name            = excluded.sport_name,
			score_state           = excluded.score_state,
			strain                = excluded.strain,
			average_heart_rate    = excluded.average_heart_rate,
			max_heart_rate        = excluded.max_heart_rate,
			kilojoule             = excluded.kilojoule,
			percent_recorded      = excluded.percent_recorded,
			distance_meter        = excluded.distance_meter,
			altitude_gain_meter   = excluded.altitude_gain_meter,
			altitude_change_meter = excluded.altitude_change_meter,
			zone_zero_milli       = excluded.zone_zero_milli,
			zone_one_milli        = excluded.zone_one_milli,
			zone_two_milli        = excluded.zone_two_milli,
			zone_three_milli      = excluded.zone_three_milli,
			zone_four_milli       = excluded.zone_four_milli,
			zone_five_milli       = excluded.zone_five_milli`

	for _, r := range records {
		var (
			sportID   any = nil
			sportName any = nil
		)
		if r.SportID != 0 {
			sportID = r.SportID
		}
		if r.SportName != "" {
			sportName = r.SportName
		}

		if _, err := tx.Exec(stmt,
			r.ID, r.UserID, formatSQLiteTime(r.CreatedAt), formatSQLiteTime(r.UpdatedAt),
			formatSQLiteTime(r.Start), formatSQLiteTime(r.End), r.TimezoneOffset,
			sportID, sportName, r.ScoreState,
			r.Score.Strain, r.Score.AverageHeartRate, r.Score.MaxHeartRate,
			r.Score.Kilojoule, r.Score.PercentRecorded, r.Score.DistanceMeter,
			r.Score.AltitudeGainMeter, r.Score.AltitudeChangeMeter,
			r.Score.ZoneDuration.ZoneZeroMilli, r.Score.ZoneDuration.ZoneOneMilli,
			r.Score.ZoneDuration.ZoneTwoMilli, r.Score.ZoneDuration.ZoneThreeMilli,
			r.Score.ZoneDuration.ZoneFourMilli, r.Score.ZoneDuration.ZoneFiveMilli,
		); err != nil {
			return fmt.Errorf("upsert workout_records id=%s: %w", r.ID, err)
		}
	}
	return nil
}
