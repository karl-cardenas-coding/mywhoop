// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Setup sets up the AWS S3 export and any resources required
func (f *AWS_S3) Setup() error {
	return nil
}

// NewAwsS3Export creates a new AWS S3 export
func NewAwsS3Export(region, bucket, profile string, client *http.Client, f *FileExport, serverMode bool) (*AWS_S3, error) {

	if region == "" {

		envValue := os.Getenv("AWS_REGION")
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

	creds, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return nil, errors.New("ERROR RETRIEVING AWS CREDENTIALS")
	}
	if creds.Expired() {
		return nil, errors.New("AWS CREDENTIALS EXPIRED")
	}

	s3Client := s3.NewFromConfig(cfg)

	err = fileExportDefaults(f)
	if err != nil {
		return nil, err
	}

	return &AWS_S3{
		Region:     region,
		Bucket:     bucket,
		S3Client:   s3Client,
		FileConfig: *f,
	}, nil
}

// Export exports the data to AWS S3
func (f *AWS_S3) Export(data []byte) error {

	if f == nil || f.S3Client == nil {
		return errors.New("s3 client is required")
	}

	s3c := f.S3Client

	err := uploadCheck(&data, &f.FileConfig, f.Bucket)
	if err != nil {
		return err
	}

	fileName := generateObjectKey(f.FileConfig)

	// If the data is greater than 5 MB part size, upload the data in parts.
	if int64(len(data)) > manager.MinUploadPartSize {

		largeBuffer := bytes.NewReader(data)
		uploader := manager.NewUploader(s3c, func(u *manager.Uploader) {
			u.PartSize = manager.MinUploadPartSize * 1024 * 1024
		})
		_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket:   &f.Bucket,
			Key:      &fileName,
			Body:     largeBuffer,
			Metadata: map[string]string{"Content-Type": "application/json"},
		})

		if err != nil {
			return err
		}

	}

	// If the data is less than 5 MB, upload the data as a single part.
	if int64(len(data)) <= manager.MinUploadPartSize {

		_, err = s3c.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:   &f.Bucket,
			Key:      &fileName,
			Body:     bytes.NewReader(data),
			Metadata: map[string]string{"Content-Type": "application/json"},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// CleanUp cleans up the AWS S3 export and any resources
func (f *AWS_S3) CleanUp() error {
	return nil
}

// fileExportDefaults sets the default values for the file export
func fileExportDefaults(f *FileExport) error {

	h, err := os.UserHomeDir()
	if err != nil {
		return errors.New("unable to get user home directory")
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

	if f != nil {

		if f.FilePath == "" {
			f.FilePath = path.Join(h, "data")
		}

		if f.FileType == "" || f.FileType != "json" {
			f.FileType = "json"
		}

	}

	return nil
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
