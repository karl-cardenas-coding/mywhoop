// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

type Export interface {
	Setup() error
	Export(data []byte) error
	CleanUp() error
}

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
}
