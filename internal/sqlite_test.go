// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func sampleUser(t *testing.T) User {
	t.Helper()
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	return User{
		UserData: UserData{
			UserID:    42,
			Email:     "rider@example.com",
			FirstName: "Ada",
			LastName:  "Lovelace",
		},
		UserMeasurements: UserMeasurements{
			HeightMeter:    1.72,
			WeightKilogram: 65.5,
			MaxHeartRate:   191,
		},
		SleepCollection: SleepCollection{
			SleepCollectionRecords: []SleepCollectionRecords{{
				ID:             "sleep-1",
				UserID:         42,
				CreatedAt:      now,
				UpdatedAt:      now,
				Start:          now.Add(-8 * time.Hour),
				End:            now,
				TimezoneOffset: "-05:00",
				Nap:            false,
				ScoreState:     "SCORED",
				Score: Score{
					StageSummary: StageSummary{
						TotalInBedTimeMilli:         28800000,
						TotalLightSleepTimeMilli:    14400000,
						TotalSlowWaveSleepTimeMilli: 7200000,
						TotalRemSleepTimeMilli:      5400000,
						SleepCycleCount:             5,
					},
					RespiratoryRate:            15.2,
					SleepPerformancePercentage: 88.0,
					SleepConsistencyPercentage: 75.0,
					SleepEfficiencyPercentage:  93.0,
				},
			}},
		},
		CycleCollection: CycleCollection{
			Records: []CycleRecords{{
				ID:             1001,
				UserID:         42,
				CreatedAt:      now,
				UpdatedAt:      now,
				Start:          now.Add(-24 * time.Hour),
				End:            now,
				TimezoneOffset: "-05:00",
				ScoreState:     "SCORED",
				Score: CycleScore{
					Strain:           12.4,
					Kilojoule:        9200.5,
					AverageHeartRate: 72,
					MaxHeartRate:     160,
				},
			}},
		},
		RecoveryCollection: RecoveryCollection{
			RecoveryRecords: []RecoveryRecords{{
				CycleID:    1001,
				SleepID:    "sleep-1",
				UserID:     42,
				CreatedAt:  now,
				UpdatedAt:  now,
				ScoreState: "SCORED",
				Score: RecoveryScore{
					UserCalibrating:  false,
					RecoveryScore:    68,
					RestingHeartRate: 54,
					HrvRmssdMilli:    42.8,
					Spo2Percentage:   97.2,
					SkinTempCelsius:  33.4,
				},
			}},
		},
		WorkoutCollection: WorkoutCollection{
			Records: []WorkoutRecords{{
				ID:             "workout-1",
				UserID:         42,
				CreatedAt:      now,
				UpdatedAt:      now,
				Start:          now.Add(-2 * time.Hour),
				End:            now.Add(-1 * time.Hour),
				TimezoneOffset: "-05:00",
				SportID:        17,
				SportName:      "cycling",
				ScoreState:     "SCORED",
				Score: WorkoutScore{
					Strain:           8.4,
					AverageHeartRate: 135,
					MaxHeartRate:     172,
					Kilojoule:        1800,
					PercentRecorded:  100,
					DistanceMeter:    20000,
					ZoneDuration: ZoneDuration{
						ZoneZeroMilli:  300000,
						ZoneOneMilli:   600000,
						ZoneTwoMilli:   900000,
						ZoneThreeMilli: 1200000,
						ZoneFourMilli:  600000,
						ZoneFiveMilli:  0,
					},
				},
			}},
		},
	}
}

func openForAssertions(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, path)
	if err != nil {
		t.Fatalf("open sqlite for assertions: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func TestSQLiteExport_SetupCreatesFileAndSchema(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(filepath.Join(dir, "nested"), "user", "")

	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = s.CleanUp() })

	if _, err := os.Stat(s.Path()); err != nil {
		t.Fatalf("sqlite file not created: %v", err)
	}

	db := openForAssertions(t, s.Path())

	wantTables := []string{
		"user_data", "user_measurements", "sleep_records", "cycle_records",
		"recovery_records", "workout_records",
	}
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var got []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, n)
	}

	sort.Strings(wantTables)
	existing := map[string]bool{}
	for _, n := range got {
		existing[n] = true
	}
	for _, want := range wantTables {
		if !existing[want] {
			t.Errorf("missing table %q; got %v", want, got)
		}
	}
}

func TestSQLiteExport_SetupCreatesExpectedIndexes(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(dir, "user", "")
	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = s.CleanUp() })

	db := openForAssertions(t, s.Path())

	wantIdx := []string{
		"idx_sleep_records_user_start",
		"idx_cycle_records_user_start",
		"idx_workout_records_user_start",
		"idx_recovery_records_user_sleep",
	}
	for _, idx := range wantIdx {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&name)
		if err != nil {
			t.Errorf("index %q not present: %v", idx, err)
		}
	}
}

func TestSQLiteExport_ExportUserWritesAllTables(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(dir, "user", "")
	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = s.CleanUp() })

	if err := s.ExportUser(sampleUser(t)); err != nil {
		t.Fatalf("ExportUser: %v", err)
	}

	db := openForAssertions(t, s.Path())
	cases := map[string]int{
		"user_data":         1,
		"user_measurements": 1,
		"sleep_records":     1,
		"cycle_records":     1,
		"recovery_records":  1,
		"workout_records":   1,
	}
	for table, want := range cases {
		if got := countRows(t, db, table); got != want {
			t.Errorf("%s row count = %d, want %d", table, got, want)
		}
	}
}

func TestSQLiteExport_Idempotent(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(dir, "user", "")
	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = s.CleanUp() })

	u := sampleUser(t)
	for i := 0; i < 3; i++ {
		if err := s.ExportUser(u); err != nil {
			t.Fatalf("ExportUser run %d: %v", i, err)
		}
	}

	db := openForAssertions(t, s.Path())
	cases := []string{"user_data", "user_measurements", "sleep_records", "cycle_records", "recovery_records", "workout_records"}
	for _, table := range cases {
		if got := countRows(t, db, table); got != 1 {
			t.Errorf("%s row count after 3 runs = %d, want 1", table, got)
		}
	}
}

func TestSQLiteExport_SurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	path := ""

	first := NewSQLiteExport(dir, "user", "")
	if err := first.Setup(); err != nil {
		t.Fatalf("Setup (first): %v", err)
	}
	if err := first.ExportUser(sampleUser(t)); err != nil {
		t.Fatalf("ExportUser (first): %v", err)
	}
	path = first.Path()
	if err := first.CleanUp(); err != nil {
		t.Fatalf("CleanUp (first): %v", err)
	}

	u2 := sampleUser(t)
	u2.SleepCollection.SleepCollectionRecords[0].ID = "sleep-2"
	u2.CycleCollection.Records[0].ID = 1002
	u2.RecoveryCollection.RecoveryRecords[0].CycleID = 1002
	u2.RecoveryCollection.RecoveryRecords[0].SleepID = "sleep-2"
	u2.WorkoutCollection.Records[0].ID = "workout-2"

	second := NewSQLiteExport(dir, "user", "")
	if err := second.Setup(); err != nil {
		t.Fatalf("Setup (second): %v", err)
	}
	t.Cleanup(func() { _ = second.CleanUp() })
	if err := second.ExportUser(u2); err != nil {
		t.Fatalf("ExportUser (second): %v", err)
	}

	if second.Path() != path {
		t.Fatalf("second run wrote to %q, want %q", second.Path(), path)
	}

	db := openForAssertions(t, path)
	if got := countRows(t, db, "sleep_records"); got != 2 {
		t.Errorf("sleep_records after two runs = %d, want 2", got)
	}
	if got := countRows(t, db, "cycle_records"); got != 2 {
		t.Errorf("cycle_records after two runs = %d, want 2", got)
	}
	if got := countRows(t, db, "recovery_records"); got != 2 {
		t.Errorf("recovery_records after two runs = %d, want 2", got)
	}
	if got := countRows(t, db, "workout_records"); got != 2 {
		t.Errorf("workout_records after two runs = %d, want 2", got)
	}
}

func TestSQLiteExport_UpdatesOnConflict(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(dir, "user", "")
	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = s.CleanUp() })

	u := sampleUser(t)
	if err := s.ExportUser(u); err != nil {
		t.Fatalf("ExportUser first: %v", err)
	}

	u.UserData.Email = "ada+updated@example.com"
	u.SleepCollection.SleepCollectionRecords[0].Score.SleepPerformancePercentage = 99.9
	if err := s.ExportUser(u); err != nil {
		t.Fatalf("ExportUser second: %v", err)
	}

	db := openForAssertions(t, s.Path())

	var email string
	if err := db.QueryRow("SELECT email FROM user_data WHERE user_id=?", u.UserData.UserID).Scan(&email); err != nil {
		t.Fatalf("query user_data: %v", err)
	}
	if email != "ada+updated@example.com" {
		t.Errorf("email = %q, want updated value", email)
	}

	var perf float64
	if err := db.QueryRow("SELECT sleep_performance_percentage FROM sleep_records WHERE id=?", "sleep-1").Scan(&perf); err != nil {
		t.Fatalf("query sleep_records: %v", err)
	}
	if perf != 99.9 {
		t.Errorf("sleep_performance_percentage = %v, want 99.9", perf)
	}
}

func TestSQLiteExport_ExportBytesDelegatesToExportUser(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(dir, "user", "")
	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = s.CleanUp() })

	payload, err := json.Marshal(sampleUser(t))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := s.Export(payload); err != nil {
		t.Fatalf("Export([]byte): %v", err)
	}

	db := openForAssertions(t, s.Path())
	if got := countRows(t, db, "user_data"); got != 1 {
		t.Errorf("user_data = %d, want 1", got)
	}
}

func TestSQLiteExport_CleanUpIdempotent(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(dir, "user", "")
	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := s.CleanUp(); err != nil {
		t.Fatalf("CleanUp first: %v", err)
	}
	if err := s.CleanUp(); err != nil {
		t.Fatalf("CleanUp second: %v", err)
	}
}

func TestSQLiteExport_ExportUserBeforeSetupErrors(t *testing.T) {
	s := &SQLiteExport{}
	if err := s.ExportUser(sampleUser(t)); err == nil {
		t.Fatalf("expected error when ExportUser is called before Setup")
	}
}

func TestSQLiteExport_NullableWorkoutFields(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(dir, "user", "")
	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = s.CleanUp() })

	u := sampleUser(t)
	u.WorkoutCollection.Records[0].SportID = 0
	u.WorkoutCollection.Records[0].SportName = ""
	if err := s.ExportUser(u); err != nil {
		t.Fatalf("ExportUser: %v", err)
	}

	db := openForAssertions(t, s.Path())
	var sportID sql.NullInt64
	var sportName sql.NullString
	if err := db.QueryRow("SELECT sport_id, sport_name FROM workout_records WHERE id=?", "workout-1").Scan(&sportID, &sportName); err != nil {
		t.Fatalf("query workout_records: %v", err)
	}
	if sportID.Valid {
		t.Errorf("sport_id expected NULL, got %d", sportID.Int64)
	}
	if sportName.Valid {
		t.Errorf("sport_name expected NULL, got %q", sportName.String)
	}
}

// copyFile copies the file at src to dst byte-for-byte. Used by the WAL
// checkpoint test to simulate "uploading just the main .sqlite file" without
// the -wal / -shm sidecars.
func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("open src: %v", err)
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create dst: %v", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		t.Fatalf("copy: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close dst: %v", err)
	}
}

// TestSQLiteExport_CheckpointFlushesWAL is the regression test for the WAL
// + S3 upload bug: in WAL mode, a small commit may live entirely in the
// <db>-wal sidecar until SQLite's auto-checkpoint threshold (~4MB) is hit.
// Copying just the main .sqlite file (which is what the S3 upload path does)
// would therefore be missing that data unless we explicitly checkpoint first.
func TestSQLiteExport_CheckpointFlushesWAL(t *testing.T) {
	dir := t.TempDir()
	s := NewSQLiteExport(dir, "user", "")
	if err := s.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = s.CleanUp() })

	if err := s.ExportUser(sampleUser(t)); err != nil {
		t.Fatalf("ExportUser: %v", err)
	}

	// Sanity check: with the DB still open and no checkpoint forced, we
	// expect the WAL sidecar to exist and hold the recent commit.
	walPath := s.Path() + "-wal"
	if info, err := os.Stat(walPath); err != nil || info.Size() == 0 {
		t.Skipf("test environment did not produce a non-empty WAL sidecar (err=%v); cannot exercise the checkpoint path", err)
	}

	if err := s.Checkpoint(); err != nil {
		t.Fatalf("Checkpoint: %v", err)
	}

	// PRAGMA wal_checkpoint(TRUNCATE) drains and zero-truncates the WAL.
	if info, err := os.Stat(walPath); err == nil && info.Size() != 0 {
		t.Errorf("WAL not truncated after checkpoint: size=%d", info.Size())
	}

	// Now copy just the main file (no -wal / -shm) and prove the data is
	// fully present in the snapshot.
	snapshot := filepath.Join(dir, "snapshot.sqlite")
	copyFile(t, s.Path(), snapshot)

	db, err := sql.Open(sqliteDriverName, snapshot)
	if err != nil {
		t.Fatalf("open snapshot: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	for _, table := range []string{"user_data", "sleep_records", "cycle_records", "recovery_records", "workout_records"} {
		if got := countRows(t, db, table); got != 1 {
			t.Errorf("snapshot %s row count = %d, want 1 (checkpoint did not flush this table)", table, got)
		}
	}
}

func TestSQLiteExport_CheckpointBeforeSetupErrors(t *testing.T) {
	s := &SQLiteExport{}
	if err := s.Checkpoint(); err == nil {
		t.Fatalf("expected error when Checkpoint is called before Setup")
	}
}

func TestSQLiteFileName(t *testing.T) {
	cases := []struct {
		prefix string
		name   string
		want   string
	}{
		{"", "", "user.sqlite"},
		{"", "mydata", "mydata.sqlite"},
		{"prod", "mydata", "prod_mydata.sqlite"},
	}
	for _, c := range cases {
		if got := sqliteFileName(c.prefix, c.name); got != c.want {
			t.Errorf("sqliteFileName(%q,%q) = %q, want %q", c.prefix, c.name, got, c.want)
		}
	}
}
