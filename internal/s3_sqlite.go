// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3SQLiteExport maintains a local cumulative sqlite database and uploads the full file
// to S3 after every successful run. The same stable object key is used on every upload -
// enable S3 versioning on the bucket if you want historical snapshots.
type S3SQLiteExport struct {
	// SQLite is the local database writer. Setup/ExportUser/CleanUp are delegated to it.
	SQLite *SQLiteExport
	// Bucket is the destination S3 bucket.
	Bucket string
	// Client is an initialized S3 client.
	Client *s3.Client
}

// NewS3SQLiteExport wires a local SQLiteExport to an S3 destination.
func NewS3SQLiteExport(sqlite *SQLiteExport, client *s3.Client, bucket string) *S3SQLiteExport {
	return &S3SQLiteExport{
		SQLite: sqlite,
		Bucket: bucket,
		Client: client,
	}
}

// Setup initializes the local sqlite database and validates the S3 destination.
func (s *S3SQLiteExport) Setup() error {
	if s.SQLite == nil {
		return errors.New("s3+sqlite exporter missing local sqlite writer")
	}
	if s.Client == nil {
		return errors.New("s3+sqlite exporter missing s3 client")
	}
	if s.Bucket == "" {
		return errors.New("s3+sqlite exporter missing bucket")
	}
	return s.SQLite.Setup()
}

// ExportUser upserts into the local sqlite file then uploads the whole file to S3.
func (s *S3SQLiteExport) ExportUser(user User) error {
	if err := s.SQLite.ExportUser(user); err != nil {
		return err
	}
	return s.uploadLocalDB(context.TODO())
}

// Export implements the Export interface by JSON-decoding into a User and delegating to ExportUser.
func (s *S3SQLiteExport) Export(data []byte) error {
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return fmt.Errorf("s3+sqlite export: decode user payload: %w", err)
	}
	return s.ExportUser(user)
}

// CleanUp closes the local database handle.
func (s *S3SQLiteExport) CleanUp() error {
	if s.SQLite == nil {
		return nil
	}
	return s.SQLite.CleanUp()
}

func (s *S3SQLiteExport) uploadLocalDB(ctx context.Context) error {
	path := s.SQLite.Path()
	if path == "" {
		return errors.New("local sqlite path is empty; was Setup called?")
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open local sqlite %q: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat local sqlite %q: %w", path, err)
	}

	key := sqliteFileName(s.SQLite.FileNamePrefix, s.SQLite.FileName)
	contentType := "application/vnd.sqlite3"

	// Transfer manager v2 picks single-part or multipart based on file size, so we
	// stream the file handle directly instead of buffering the whole db in memory.
	tm := transfermanager.New(s.Client)
	if _, err := tm.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      &s.Bucket,
		Key:         &key,
		Body:        f,
		ContentType: &contentType,
	}); err != nil {
		return fmt.Errorf("upload sqlite to s3://%s/%s: %w", s.Bucket, key, err)
	}
	slog.Info("sqlite uploaded to s3", "bucket", s.Bucket, "key", key, "bytes", info.Size())
	return nil
}
