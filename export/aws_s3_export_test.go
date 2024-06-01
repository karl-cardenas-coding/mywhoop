package export

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"
)

func TestNewAwsS3Export(t *testing.T) {

	client := http.DefaultClient

	tests := []struct {
		id            int
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
			0,
			false,
			"us-east-1",
			"mywhoop",
			"",
			&FileExport{},
			client,
			false,
			false,
			false,
		},
		// Happy path with profile ENV variable
		{
			0,
			false,
			"us-east-1",
			"mywhoop",
			"",
			&FileExport{},
			client,
			true,
			false,
			false,
		},
		// Error case: missing region
		{
			0,
			false,
			"",
			"mywhoop",
			"",
			&FileExport{},
			client,
			false,
			false,
			true,
		},
		// happy path with region ENV variable
		{
			0,
			false,
			"",
			"mywhoop",
			"",
			&FileExport{},
			client,
			false,
			true,
			false,
		},
		// Error case: missing bucket
		{
			0,
			false,
			"us-east-1",
			"",
			"",
			&FileExport{},
			client,
			false,
			false,
			true,
		},
		// Happy path: server mode enabled
		{
			0,
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

	for index, tc := range tests {
		tc.id = index + 1

		if tc.setProfileEnv {
			os.Setenv("AWS_PROFILE", "test")
		}

		if tc.setRegionEnv {
			os.Setenv("AWS_REGION", "us-east-1")
		}

		result, err := NewAwsS3Export(tc.region, tc.bucket, tc.profile, tc.client, tc.f, tc.serverMode)
		if err != nil && !tc.expectedError {
			t.Errorf("Test Case - %d: Unexpected error: %v", tc.id, err)
		}

		if err == nil && tc.expectedError {
			t.Errorf("Test Case - %d: Expected error, but got no error", tc.id)
		}

		if err == nil && result.S3Client == nil {
			t.Errorf("Test Case - %d: S3 client is nil", tc.id)
		}

		if err == nil && result.FileConfig.ServerMode != true {
			t.Errorf("Test Case - %d: Server mode is not set correctly", tc.id)
		}

		clearEnvVariables()
	}

}

func TestFileExportDefaults(t *testing.T) {

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
				FileName:       "user",
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
				FileName:       "user",
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
	os.Unsetenv("AWS_REGION")
}
