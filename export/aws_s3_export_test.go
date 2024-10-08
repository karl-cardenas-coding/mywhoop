// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/docker/go-connections/nat"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestNewAwsS3Export(t *testing.T) {

	client := http.DefaultClient

	tests := []struct {
		description   string
		serverMode    bool
		region        string
		bucket        string
		profile       string
		f             *FileExport
		client        *http.Client
		setProfileEnv bool
		setRegionEnv  bool
		expectedError bool
	}{
		// Happy path
		{
			"Test case 1: Happy path",
			false,
			"us-east-1",
			"mywhoop",
			"",
			&FileExport{
				ServerMode: true,
			},
			client,
			false,
			false,
			false,
		},
		// Happy path with profile ENV variable
		{
			"Test case 2: Happy path with profile ENV variable",
			false,
			"us-east-1",
			"mywhoop",
			"",
			&FileExport{
				ServerMode: true,
			},
			client,
			false,
			false,
			false,
		},
		// Error case: missing region
		{
			"Test case 3: Error case: missing region",
			false,
			"",
			"mywhoop",
			"",
			&FileExport{
				ServerMode: true,
			},
			client,
			false,
			false,
			true,
		},
		// happy path with region ENV variable
		{
			"Test case 4: Happy path with region ENV variable",
			false,
			"",
			"mywhoop",
			"",
			&FileExport{
				ServerMode: true,
			},
			client,
			false,
			true,
			false,
		},
		// Error case: missing bucket
		{
			"Test case 5: Error case: missing bucket",
			false,
			"us-east-1",
			"",
			"",
			&FileExport{
				ServerMode: true,
			},
			client,
			false,
			false,
			true,
		},
		// Happy path: server mode enabled
		{
			"Test case 6: Happy path: server mode enabled",
			true,
			"us-east-1",
			"mywhoop",
			"",
			&FileExport{
				ServerMode: true,
			},
			client,
			false,
			false,
			false,
		},
	}

	for _, tc := range tests {

		// if tc.clearEnv {

		// 	t.Setenv("AWS_PROFILE", "")
		// 	t.Setenv("AWS_DEFAULT_REGION", "")
		// 	t.Setenv("AWS_ACCESS_KEY_ID", "")
		// 	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
		// }

		t.Run(tc.description, func(t *testing.T) {

			if tc.setProfileEnv {
				os.Setenv("AWS_PROFILE", "test")
			}

			os.Unsetenv("AWS_DEFAULT_REGION")

			if tc.setRegionEnv {
				os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
			}

			result, err := NewAwsS3Export(tc.region, tc.bucket, tc.profile, tc.client, tc.f, tc.serverMode)
			if err != nil && !tc.expectedError {
				t.Errorf("%s: Unexpected error: %v", tc.description, err)
			}

			if err == nil && tc.expectedError {
				t.Errorf("%s: Expected error, but got no error", tc.description)
			}

			if err == nil && result.S3Client == nil {
				t.Errorf("%s: S3 client is nil", tc.description)
			}

			if err == nil && result.FileConfig.ServerMode != true {
				t.Errorf("%s: Server mode is not set correctly", tc.description)
			}

		})
		t.Cleanup(func() {
			clearEnvVariables()
		})

	}

}

func TestAWSS3Export(t *testing.T) {

	ctx := context.Background()
	newNetwork, err := network.New(ctx)
	if err != nil {
		t.Errorf("failed to create network: %s", err)
	}

	localstackContainer, err := localstack.Run(ctx,
		"localstack/localstack:3.5",
		testcontainers.WithEnv(map[string]string{
			"SERVICES": "s3"}),
		network.WithNetwork([]string{"localstack-network-v2"}, newNetwork),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// Create an S3 client
	s3Client, err := s3Client(ctx, localstackContainer, "us-east-1")
	if err != nil {
		t.Fatalf("failed to create S3 client: %s", err)

	}

	mockBucketName := "mywhoop"

	err = createMockBucket(s3Client, mockBucketName)
	if err != nil {
		t.Fatalf("failed to create mock bucket: %s", err)
	}

	tests := []struct {
		description   string
		awsS3         *AWS_S3
		data          []byte
		errorExpected bool
	}{
		{
			description: "Test case 1: Export data to S3",
			awsS3: &AWS_S3{
				Region:   "us-east-1",
				Bucket:   mockBucketName,
				S3Client: s3Client,
				FileConfig: FileExport{
					FilePath:       "/tmp",
					FileType:       "json",
					FileName:       "user",
					FileNamePrefix: "",
					ServerMode:     true,
				},
			},
			data:          []byte(`{"name": "John Doe", "age": 30}`),
			errorExpected: false,
		},
		{
			description: "Test case 2: Empty data",
			awsS3: &AWS_S3{
				Region:   "us-east-1",
				Bucket:   mockBucketName,
				S3Client: s3Client,
				FileConfig: FileExport{
					FilePath:       "/tmp",
					FileType:       "json",
					FileName:       "user",
					FileNamePrefix: "",
					ServerMode:     true,
				},
			},
			data:          nil,
			errorExpected: true,
		},
		{
			description: "Test case 3: 0 data length",
			awsS3: &AWS_S3{
				Region:   "us-east-1",
				Bucket:   mockBucketName,
				S3Client: s3Client,
				FileConfig: FileExport{
					FilePath:       "/tmp",
					FileType:       "json",
					FileName:       "user",
					FileNamePrefix: "",
					ServerMode:     true,
				},
			},
			data:          []byte(""),
			errorExpected: true,
		},
		{
			description:   "Test case 4: Nil AWS S3",
			awsS3:         nil,
			data:          []byte(`{"name": "John Doe", "age": 30}`),
			errorExpected: true,
		},
		{
			description: "Test case 5: over 10 MB data",
			awsS3: &AWS_S3{
				Region:   "us-east-1",
				Bucket:   mockBucketName,
				S3Client: s3Client,
				FileConfig: FileExport{
					FilePath:       "/tmp",
					FileType:       "json",
					FileName:       "user",
					FileNamePrefix: "",
					ServerMode:     true,
				},
			},
			data:          generateRandomData(10 * 1024 * 1024),
			errorExpected: false,
		},
		{
			description: "Test case 6: No bucket",
			awsS3: &AWS_S3{
				Region:   "us-east-1",
				Bucket:   "",
				S3Client: s3Client,
				FileConfig: FileExport{
					FilePath:       "/tmp",
					FileType:       "json",
					FileName:       "user",
					FileNamePrefix: "",
					ServerMode:     true,
				},
			},
			data:          []byte(`{"name": "John Doe", "age": 30}`),
			errorExpected: true,
		},
	}

	for index, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.awsS3.Export(tc.data)
			if !tc.errorExpected && err != nil {
				t.Errorf("Test Case - %d: Unexpected error: %v", index+1, err)
			}

			if tc.errorExpected && err == nil {
				t.Errorf("Test Case - %d: Expected error, but got no error", index+1)
			}
		})
	}

}

func TestFileAWSS3ExportDefaults(t *testing.T) {

	h, err := os.UserHomeDir()
	if err != nil {
		t.Errorf("Unable to get user home directory")
	}

	homePath := path.Join(h, "data")

	tests := []struct {
		id          int
		description string
		file        *FileExport
		expected    *FileExport
	}{
		{
			0,
			"Test case 1: File export with empty file export",
			&FileExport{},
			&FileExport{
				FilePath:       homePath,
				FileType:       "json",
				FileName:       "",
				FileNamePrefix: "",
				ServerMode:     true,
			},
		}, {

			0,
			"Test case 2: File export with custom file path",
			&FileExport{
				FilePath:       "/tmp",
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			&FileExport{
				FilePath:       "/tmp",
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
		},
		{
			0,
			"Test case 3: File export with server mode disabled",
			&FileExport{
				FilePath:       homePath,
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     false,
			},
			&FileExport{
				FilePath:       homePath,
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     false,
			},
		},
		{
			0,
			"Test case 4: File export with invalid file type",
			&FileExport{
				FilePath:       homePath,
				FileType:       "csv",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			&FileExport{
				FilePath:       homePath,
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
		},
		{
			0,
			"Test case 5: File export with empty file name",
			&FileExport{
				FilePath:       homePath,
				FileType:       "json",
				FileName:       "",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			&FileExport{
				FilePath:       homePath,
				FileType:       "json",
				FileName:       "",
				FileNamePrefix: "",
				ServerMode:     true,
			},
		},
		{
			0,
			"Test case 6: Nil file export",
			nil,
			&FileExport{
				FilePath:       homePath,
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
		},
		{

			0,
			"Test case 7: File export with Excel file type",
			&FileExport{
				FilePath:       "/tmp",
				FileType:       "xlsx",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			&FileExport{
				FilePath:       "/tmp",
				FileType:       "xlsx",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
		},
	}

	for index, tc := range tests {
		tc.id = index + 1
		file := tc.file
		err := fileExportDefaults(file)
		if err != nil {
			t.Errorf("%s: Unexpected error: %v", tc.description, err)
		}

		if file == nil {

			if tc.expected.FilePath != homePath {
				t.Errorf("%s: Expected file path %s, got %s", tc.description, homePath, tc.expected.FilePath)
			}

			if tc.expected.FileName != "user" {
				t.Errorf("%s: Expected file name user, got %s", tc.description, tc.expected.FileName)
			}

			if tc.expected.FileType != "json" {
				t.Errorf("%s: Expected file type json, got %s", tc.description, tc.expected.FileType)
			}

			if tc.expected.FileNamePrefix != "" {
				t.Errorf("%s: Expected file name prefix empty, got %s", tc.description, tc.expected.FileNamePrefix)
			}

			if tc.expected.ServerMode != true {
				t.Errorf("%s: Expected server mode true, got %v", tc.description, tc.expected.ServerMode)
			}

		}

		if file != nil {
			if file.FilePath != tc.expected.FilePath {
				t.Errorf("%s: Expected file path %s, got %s", tc.description, tc.expected.FilePath, file.FilePath)
			}

			if file.FileType != tc.expected.FileType {
				t.Errorf("%s: Expected file type %s, got %s", tc.description, tc.expected.FileType, file.FileType)
			}

			if file.FileName != tc.expected.FileName {
				t.Errorf("%s: Expected file name %s, got %s", tc.description, tc.expected.FileName, file.FileName)
			}

			if file.FileNamePrefix != tc.expected.FileNamePrefix {
				t.Errorf("%s: Expected file name prefix %s, got %s", tc.description, tc.expected.FileNamePrefix, file.FileNamePrefix)
			}
		}

	}
}

func TestGenerateObjectKeye(t *testing.T) {

	type test struct {
		testCase    int
		description string
		file        FileExport
		want        string
	}

	tests := []test{
		{
			description: "Test case 1: File name with custom prefix prefix",
			file: FileExport{
				FileNamePrefix: "test",
				FileName:       "user",
				FileType:       "json",
				ServerMode:     false,
			},
			want: "test_user.json",
		},
		{
			description: "Test case 2: File name with empty prefix",
			file: FileExport{
				FileNamePrefix: "",
				FileName:       "user",
				FileType:       "json",
				ServerMode:     false,
			},
			want: "user.json",
		},
		{
			description: "Test case 3: File name with empty prefix and file name",
			file: FileExport{
				FileNamePrefix: "",
				FileName:       "",
				FileType:       "",
				ServerMode:     false,
			},
			want: ".",
		},
		{
			description: "Test case 4: Server mode enabled",
			file: FileExport{
				FileNamePrefix: "",
				FileName:       "user",
				FileType:       "json",
				ServerMode:     true,
			},
			want: fmt.Sprintf("user_%s.json", getCurrentDate()),
		},
		{
			description: "Test case 5: Server mode enabled with custom prefix",
			file: FileExport{
				FileNamePrefix: "test",
				FileName:       "user",
				FileType:       "json",
				ServerMode:     true,
			},
			want: fmt.Sprintf("test_user_%s.json", getCurrentDate()),
		},
	}

	for index, tc := range tests {
		tc.testCase = index
		got := generateObjectKey(tc.file)
		if got != tc.want {
			t.Errorf("%s - Expected %s error, got: %v", tc.description, tc.want, got)
		}
	}

}

func TestS3SetUp(t *testing.T) {

}

func TestS3Export(t *testing.T) {

}

func TestS3CleanUp(t *testing.T) {

}

func clearEnvVariables() {
	os.Unsetenv("AWS_PROFILE")
	os.Unsetenv("AWS_DEFAULT_REGION")
}

func createMockBucket(s3Client *s3.Client, bucket string) error {

	_, err := s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		return err
	}

	return nil

}

func s3Client(ctx context.Context, l *localstack.LocalStackContainer, region string) (*s3.Client, error) {
	mappedPort, err := l.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		return nil, err
	}

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return nil, err
	}
	defer provider.Close()

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return nil, err
	}

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("aaaa", "bbb", "cccc")),
	)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = &endpoint
		o.UsePathStyle = true
	})

	return client, nil
}

func generateRandomData(size int) []byte {

	r := rand.New(rand.NewPCG(1, 2))

	data := make([]byte, size)
	for i := range data {
		data[i] = byte(r.Int())
	}
	fmt.Println("Generated Data with a length of: ", len(data))
	return data
}

func TestAWSS3Setup(t *testing.T) {
	awsS3 := &AWS_S3{}
	err := awsS3.Setup()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAWSS3Cleanup(t *testing.T) {
	awsS3 := &AWS_S3{}
	err := awsS3.CleanUp()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestUploadCheck(t *testing.T) {

	tests := []struct {
		description   string
		data          []byte
		f             *FileExport
		bucket        string
		errorExpected bool
	}{
		{
			description: "Test case 1: Happy path",
			data:        []byte(`{"name": "John Doe", "age": 30}`),
			f: &FileExport{
				FilePath:       "/tmp",
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			bucket:        "mywhoop",
			errorExpected: false,
		},
		{
			description: "Test case 2: data is nil",
			data:        nil,
			f: &FileExport{
				FilePath:       "/tmp",
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			bucket:        "mywhoop",
			errorExpected: true,
		},
		{
			description: "Test case 3: data empty",
			data:        []byte(""),
			f: &FileExport{
				FilePath:       "/tmp",
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			bucket:        "mywhoop",
			errorExpected: true,
		},
		{
			description:   "Test case 4: file export is nil",
			data:          []byte(`{"name": "John Doe", "age": 30}`),
			f:             nil,
			bucket:        "mywhoop",
			errorExpected: true,
		},
		{
			description: "Test case 5: bucket is empty",
			data:        []byte(`{"name": "John Doe", "age": 30}`),
			f: &FileExport{
				FilePath:       "/tmp",
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			bucket:        "",
			errorExpected: true,
		},
		{
			description: "Test case 6: file export with empty file name",
			data:        []byte(`{"name": "John Doe", "age": 30}`),
			f: &FileExport{
				FilePath:       "/tmp",
				FileType:       "json",
				FileName:       "",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			bucket:        "mywhoop",
			errorExpected: false,
		},
		{
			description: "Test case 7: file export with empty file type",
			data:        []byte(`{"name": "John Doe", "age": 30}`),
			f: &FileExport{
				FilePath:       "/tmp",
				FileType:       "",
				FileName:       "user",
				FileNamePrefix: "",
				ServerMode:     true,
			},
			bucket:        "mywhoop",
			errorExpected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := uploadCheck(&tc.data, tc.f, tc.bucket)
			if !tc.errorExpected && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.errorExpected && err == nil {
				t.Errorf("Expected error, but got no error")
			}
		})
	}
}
