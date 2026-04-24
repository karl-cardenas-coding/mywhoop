// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// sqliteDriverName is the driver name registered by modernc.org/sqlite.
const sqliteDriverName = "sqlite"

// sqliteBusyTimeoutMS controls how long SQLite waits on a locked DB before returning SQLITE_BUSY.
const sqliteBusyTimeoutMS = 5000

// defaultSQLiteFileName is the default name of the sqlite database file (without extension).
const defaultSQLiteFileName = "user"

// SQLiteExport writes Whoop data into a single SQLite file that accumulates across runs.
// It implements both Export and UserExporter. The UserExporter path is the primary one;
// Export([]byte) is provided for callers that only have the JSON-serialized form.
type SQLiteExport struct {
	// FilePath is the folder the sqlite file lives in.
	FilePath string
	// FileName is the base name (without extension). Defaults to "user".
	FileName string
	// FileNamePrefix is an optional prefix separated by "_".
	FileNamePrefix string

	db   *sql.DB
	path string
}

// NewSQLiteExport constructs a SQLiteExport. Setup() must be called before use.
func NewSQLiteExport(filePath, fileName, fileNamePrefix string) *SQLiteExport {
	return &SQLiteExport{
		FilePath:       filePath,
		FileName:       fileName,
		FileNamePrefix: fileNamePrefix,
	}
}

// Path returns the resolved on-disk path of the sqlite file. Only valid after Setup().
func (s *SQLiteExport) Path() string {
	return s.path
}

// Setup resolves the destination path, creates the folder if needed, opens the database,
// applies pragmas, and creates the schema if it doesn't already exist.
func (s *SQLiteExport) Setup() error {
	if s.FilePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve current directory: %w", err)
		}
		s.FilePath = filepath.Join(cwd, "data")
	}

	if s.FileName == "" {
		s.FileName = defaultSQLiteFileName
	}

	if err := os.MkdirAll(s.FilePath, 0o755); err != nil {
		return fmt.Errorf("create data folder %q: %w", s.FilePath, err)
	}

	s.path = filepath.Join(s.FilePath, sqliteFileName(s.FileNamePrefix, s.FileName))

	db, err := sql.Open(sqliteDriverName, s.path)
	if err != nil {
		return fmt.Errorf("open sqlite at %q: %w", s.path, err)
	}

	// Pin the pool to a single connection. Several of the PRAGMAs below
	// (busy_timeout, synchronous, foreign_keys) are scoped to a connection,
	// so a pool that hands out fresh connections would silently lose them.
	// SQLite is also a single-writer store, so a multi-connection pool buys
	// us nothing and creates lock contention.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := configureSQLiteConnection(db); err != nil {
		_ = db.Close()
		return err
	}

	if err := applySQLiteSchema(db); err != nil {
		_ = db.Close()
		return err
	}

	s.db = db
	slog.Info("sqlite export ready", "file", s.path)
	return nil
}

// ExportUser upserts the Whoop user payload into the sqlite database inside a single
// transaction. Re-running with the same data is idempotent; re-running with new data
// extends the dataset (append-or-update by primary key).
func (s *SQLiteExport) ExportUser(user User) error {
	if s.db == nil {
		return errors.New("sqlite exporter not initialized; call Setup first")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin sqlite transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := upsertUser(tx, user); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit sqlite transaction: %w", err)
	}
	committed = true

	slog.Info("sqlite export complete", "file", s.path)
	return nil
}

// Export satisfies the Export interface by JSON-decoding the payload back into a User
// and delegating to ExportUser. Prefer calling ExportUser directly.
func (s *SQLiteExport) Export(data []byte) error {
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return fmt.Errorf("sqlite export: decode user payload: %w", err)
	}
	return s.ExportUser(user)
}

// Checkpoint forces any pending WAL frames to be merged into the main database
// file and truncates the WAL. Call this before copying or uploading the .sqlite
// file out-of-band, otherwise recent commits may still live in the -wal sidecar
// and the snapshot will be stale.
func (s *SQLiteExport) Checkpoint() error {
	if s.db == nil {
		return errors.New("sqlite exporter not initialized; call Setup first")
	}
	// wal_checkpoint(TRUNCATE) blocks until all readers/writers release the WAL,
	// then drains every frame into the main DB and zeroes the WAL file.
	if _, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return fmt.Errorf("checkpoint sqlite WAL: %w", err)
	}
	return nil
}

// CleanUp closes the database handle. It is safe to call multiple times.
func (s *SQLiteExport) CleanUp() error {
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	if err != nil {
		return fmt.Errorf("close sqlite database: %w", err)
	}
	return nil
}

// sqliteFileName composes the on-disk file name. Unlike json/xlsx exports, the sqlite
// database is not dated per run - it is a cumulative store.
func sqliteFileName(prefix, name string) string {
	if name == "" {
		name = defaultSQLiteFileName
	}
	if prefix != "" {
		return prefix + "_" + name + ".sqlite"
	}
	return name + ".sqlite"
}

func configureSQLiteConnection(db *sql.DB) error {
	pragmas := []string{
		fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMS),
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("configure sqlite (%s): %w", p, err)
		}
	}
	return nil
}

func formatSQLiteTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func boolToSQLiteInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
