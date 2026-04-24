// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"slices"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Setup sets up the AWS S3 export and any resources required
func (f *AWS_S3) Setup() error {
	return nil
}

// NewAwsS3Export creates a new AWS S3 export
func NewAwsS3Export(region, bucket, profile string, client *http.Client, f *FileExport, serverMode bool) (*AWS_S3, error) {

	if region == "" {

		envValue := os.Getenv("AWS_DEFAULT_REGION")
		if envValue == "" {
			return nil, fmt.Errorf("AWS region is required")
		}
	}

	if bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithHTTPClient(client),
		config.WithAssumeRoleCredentialOptions(func(aro *stscreds.AssumeRoleOptions) {
			aro.TokenProvider = stscreds.StdinTokenProvider
		}),
	}

	// Check if the profile is set through the ENV variable
	envValue := os.Getenv("AWS_PROFILE")
	if envValue != "" {
		// override the incoming parameter due to precedence
		profile = envValue
	}

	// If a profile is set, either through ENV var or from config, load the configuration with the profile
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return nil, fmt.Errorf("ERROR LOADING AWS CONFIG: %v", err)
	}

	if cfg.Credentials == nil {
		return nil, errors.New("ERROR LOADING AWS CREDENTIALS")

	}

	creds, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		slog.Debug("error retrieving aws credentials", "error", err)
		return nil, errors.New("ERROR RETRIEVING AWS CREDENTIALS")
	}

	if creds.Expired() {
		return nil, errors.New("AWS CREDENTIALS EXPIRED")
	}

	s3Client := s3.NewFromConfig(cfg)

	fileCfg, err := fileExportDefaults(f)
	if err != nil {
		return nil, err
	}

	return &AWS_S3{
		Region:     region,
		Bucket:     bucket,
		S3Client:   s3Client,
		FileConfig: *fileCfg,
	}, nil
}

// Export exports the data to AWS S3
func (f *AWS_S3) Export(data []byte) error {

	if f == nil || f.S3Client == nil {
		return errors.New("s3 client is required")
	}

	err := uploadCheck(&data, &f.FileConfig, f.Bucket)
	if err != nil {
		return err
	}

	fileName := generateObjectKey(f.FileConfig)
	contentType := determineContentType(f.FileConfig.FileType)

	// Transfer manager v2 transparently picks single-part vs multipart based on
	// the configured MultipartUploadThreshold, so we no longer branch on size.
	tm := transfermanager.New(f.S3Client)
	_, err = tm.UploadObject(context.TODO(), &transfermanager.UploadObjectInput{
		Bucket:      &f.Bucket,
		Key:         &fileName,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return err
	}

	return nil
}

// CleanUp cleans up the AWS S3 export and any resources
func (f *AWS_S3) CleanUp() error {
	return nil
}

// fileExportDefaults returns a populated *FileExport for use by the S3 exporter.
// Passing nil yields a fresh struct with sensible defaults; passing an existing
// struct fills in any zero-valued fields. The returned pointer is always non-nil
// on a nil error, and is always safe for the caller to dereference.
//
// Note: the previous in-place version mutated *f and silently swapped in a
// local default when nil was passed. The local-swap never reached the caller,
// so dereferencing the original nil pointer in NewAwsS3Export would crash.
// Returning the value explicitly fixes that footgun.
func fileExportDefaults(f *FileExport) (*FileExport, error) {

	supportedFileTypes := []string{"json", "xlsx", "sqlite"}

	h, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.New("unable to get user home directory")
	}

	if f == nil {
		f = &FileExport{
			FilePath:       path.Join(h, "data"),
			FileName:       "user",
			FileType:       "json",
			FileNamePrefix: "",
			ServerMode:     true,
		}
	}

	if f.FilePath == "" {
		f.FilePath = path.Join(h, "data")
	}

	if f.FileType == "" {
		f.FileType = "json"
	}

	if !slices.Contains(supportedFileTypes, f.FileType) {
		f.FileType = "json"
	}

	return f, nil
}

// generateName generates the name of the file to be created
func generateObjectKey(cfg FileExport) string {

	if cfg.ServerMode {

		if cfg.FileNamePrefix != "" && cfg.FileName != "" {
			return cfg.FileNamePrefix + "_" + cfg.FileName + "_" + getCurrentDate() + "." + cfg.FileType
		}

		if cfg.FileName != "" {
			return cfg.FileName + "_" + getCurrentDate() + "." + cfg.FileType
		}

		return getCurrentDate() + "." + cfg.FileType
	}

	if cfg.FileNamePrefix != "" {
		return cfg.FileNamePrefix + "_" + cfg.FileName + "." + cfg.FileType
	}

	if cfg.FileName != "" {
		return "user" + "." + cfg.FileType
	}

	return cfg.FileName + "." + cfg.FileType

}

// uploadCheck checks if the data, file export and bucket are valid
func uploadCheck(data *[]byte, f *FileExport, bucket string) error {

	if f == nil {
		return errors.New("file export is required")
	}

	if data == nil {
		return errors.New("data is required")

	}

	if len(*data) == 0 {
		return errors.New("data is empty")
	}

	if f.FileType == "" {
		return errors.New("file type is required")
	}

	if bucket == "" {
		return errors.New("bucket is required")
	}

	return nil
}

// determineContentType determines the content type based on the file type
func determineContentType(fileType string) string {

	switch fileType {
	case "json":
		return "application/json"
	case "csv":
		return "text/csv"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "sqlite":
		return "application/vnd.sqlite3"
	default:
		return "application/json"
	}
}
