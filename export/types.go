// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import "github.com/aws/aws-sdk-go-v2/service/s3"

type FileExport struct {
	// FilePath is the path to the file to be created. If not provided, the default path is the data folder in the current directory.
	FilePath string `yaml:"filePath"`
	// FileType is the type of file to be created. If not provided, the default type is json.
	FileType string `yaml:"fileType"`
	// FileName is the name of the file to be created. If not provided, the default name is user.
	FileName string `yaml:"fileName"`
	// FileNamePrefix is used to prefix the file name. If not provided, the default prefix is empty.
	FileNamePrefix string `yaml:"fileNamePrefix"`
}

type AWS_S3 struct {
	// The AWS region the S3 bucket is located in.
	Region string `yaml:"region"`
	// Bucket is the name of the S3 bucket.
	Bucket string `yaml:"bucket"`
	// S3Client is the S3 client.
	S3Client *s3.Client `yaml:"-"`
	// FileConfig contains the file configuration.
	FileConfig FileExport `yaml:"fileConfig"`
	// Profile is AWS profile to use.
	Profile string `yaml:"profile"`
}
