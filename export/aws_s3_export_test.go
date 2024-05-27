package export

import (
	"net/http"
	"os"
	"testing"
)

func TestNewAwsS3Export(t *testing.T) {

	client := http.DefaultClient

	tests := []struct {
		id            int
		region        string
		bucket        string
		profile       string
		client        *http.Client
		setProfileEnv bool
		setRegionEnv  bool
		expectedError bool
	}{
		// Happy path
		{
			0,
			"us-east-1",
			"mywhoop",
			"",
			client,
			false,
			false,
			false,
		},
		// Happy path with profile ENV variable
		{
			0,
			"us-east-1",
			"mywhoop",
			"",
			client,
			true,
			false,
			false,
		},
		// Error case: missing region
		{
			0,
			"",
			"mywhoop",
			"",
			client,
			false,
			false,
			true,
		},
		// happy path with region ENV variable
		{
			0,
			"",
			"mywhoop",
			"",
			client,
			false,
			true,
			false,
		},
		// Error case: missing bucket
		{
			0,
			"us-east-1",
			"",
			"",
			client,
			false,
			false,
			true,
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

		result, err := NewAwsS3Export(tc.region, tc.bucket, tc.profile, tc.client)
		if err != nil && !tc.expectedError {
			t.Errorf("Test Case - %d: Unexpected error: %v", tc.id, err)
		}

		if err == nil && tc.expectedError {
			t.Errorf("Test Case - %d: Expected error, but got no error", tc.id)
		}

		if err == nil && result.S3Client == nil {
			t.Errorf("Test Case - %d: S3 client is nil", tc.id)
		}

		clearEnvVariables()
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
