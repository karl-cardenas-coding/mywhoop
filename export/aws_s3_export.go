// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

// Setup sets up the AWS S3 export and any resources required
func (f *AWS_S3) Setup() error {
	//  TODO: Implement this method
	return nil
}

// NewAwsS3Export creates a new AWS S3 export
func NewAwsS3Export(region, bucket string) *AWS_S3 {
	return &AWS_S3{
		Region: region,
		Bucket: bucket,
	}
}

// Export exports the data to AWS S3
func (f *AWS_S3) Export(data []byte) error {
	//  TODO: Implement this method
	return nil
}

// CleanUp cleans up the AWS S3 export and any resources
func (f *AWS_S3) CleanUp() error {
	//  TODO: Implement this method
	return nil
}
