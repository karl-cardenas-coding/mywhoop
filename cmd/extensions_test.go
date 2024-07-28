// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"net/http"
	"os"
	"testing"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/karl-cardenas-coding/mywhoop/notifications"
)

func TestDetermineExporterExtension(t *testing.T) {

	client := internal.CreateHTTPClient()

	tests := []struct {
		name          string
		cfg           internal.ConfigurationData
		dataLocation  string
		client        *http.Client
		expextedError bool
		expectedType  interface{}
		setEnvCreds   bool
		setAWScreds   bool
	}{
		{
			name:         "file with datalocation",
			cfg:          internal.ConfigurationData{},
			dataLocation: "data/",
			expectedType: &export.FileExport{
				FileType:       "json",
				FileName:       "user",
				FileNamePrefix: "test_",
			},
			expextedError: false,
		},
		{
			name:         "file",
			dataLocation: "",
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					Method: "file",
					FileExport: export.FileExport{
						FilePath:       "data/",
						FileType:       "json",
						FileName:       "user",
						FileNamePrefix: "test_",
					},
				},
			},
			expectedType:  &export.FileExport{},
			expextedError: false,
		},
		{
			name:          "aws",
			expextedError: false,
			dataLocation:  "",
			setAWScreds:   true,
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					Method: "s3",
					AWSS3: export.AWS_S3{
						Region:     "us-west-2",
						Bucket:     "mybucket",
						FileConfig: export.FileExport{},
					},
				},
			},
		},
		{
			name:          "aws with datalocation",
			expextedError: false,
			dataLocation:  "whoopdata",
			setAWScreds:   true,
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					Method: "s3",
					AWSS3: export.AWS_S3{
						Region:     "us-west-2",
						Bucket:     "mybucket",
						FileConfig: export.FileExport{},
					},
				},
			},
		},
		{
			name:          "aws with error",
			expextedError: true,
			setEnvCreds:   false,
			setAWScreds:   false,
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					Method: "s3",
					AWSS3: export.AWS_S3{
						FileConfig: export.FileExport{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setEnvCreds {
				setEnvCreds(false, false, test.setAWScreds)
			}

			test.client = client

			exporterMethod, err := determineExporterExtension(test.cfg, test.client, dataLocation)
			if (err != nil) != test.expextedError {
				t.Errorf("expected error: %v, got: %v", test.expextedError, err)
			}

			// check if the returned type is the expected type
			if test.expectedType != nil {
				if _, ok := exporterMethod.(*export.FileExport); ok {
					if _, ok := test.expectedType.(*export.FileExport); !ok {
						t.Errorf("%s - expected type: %T, got: %T", test.name, test.expectedType, exporterMethod)
					}
				}

				if _, ok := exporterMethod.(*export.AWS_S3); ok {
					if _, ok := test.expectedType.(*export.AWS_S3); !ok {
						t.Errorf("%s - expected type: %T, got: %T", test.name, test.expectedType, exporterMethod)
					}
				}
			}

		})
		t.Cleanup(func() {

			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			os.Unsetenv("AWS_DEFAULT_REGION")

		})
	}
}

func TestDetermineNotificationExtension(t *testing.T) {

	tests := []struct {
		name          string
		cfg           internal.ConfigurationData
		expextedError bool
		setEnvCreds   bool
		setToken      bool
		setPassword   bool
		expectedType  interface{}
	}{
		{
			name: "ntfy",
			cfg: internal.ConfigurationData{
				Notification: internal.NotificationConfig{
					Method: "ntfy",
					Ntfy: notifications.Ntfy{
						ServerEndpoint: "http://localhost:8080",
						SubscriptionID: "1234",
						Events:         "all",
					},
				},
			},
			expextedError: false,
			setEnvCreds:   true,
			setToken:      true,
			expectedType:  &notifications.Ntfy{},
		},
		{
			name: "ntfy with error",
			cfg: internal.ConfigurationData{
				Notification: internal.NotificationConfig{
					Method: "ntfy",
					Ntfy: notifications.Ntfy{
						Events: "all",
					},
				},
			},
			expextedError: true,
			setEnvCreds:   true,
			setToken:      true,
			expectedType:  &notifications.Ntfy{},
		},
		{
			name: "no notification method specified",
			cfg: internal.ConfigurationData{
				Notification: internal.NotificationConfig{},
			},
			expextedError: false,
			expectedType:  &notifications.Stdout{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setEnvCreds {
				setEnvCreds(test.setPassword, test.setToken, false)
			}
			notificationMethod, err := determineNotificationExtension(test.cfg)
			if (err != nil) != test.expextedError {
				t.Errorf("expected error: %v, got: %v", test.expextedError, err)
			}

			// check if the returned type is the expected type
			if test.expectedType != nil {
				if _, ok := notificationMethod.(*notifications.Ntfy); ok {
					if _, ok := test.expectedType.(*notifications.Ntfy); !ok {
						t.Errorf("expected type: %T, got: %T", test.expectedType, notificationMethod)
					}
				}

				if _, ok := notificationMethod.(*notifications.Stdout); ok {
					if _, ok := test.expectedType.(*notifications.Stdout); !ok {
						t.Errorf("expected type: %T, got: %T", test.expectedType, notificationMethod)
					}
				}
			}

		})
		t.Cleanup(func() {
			if test.setEnvCreds {
				os.Unsetenv("NOTIFICATION_NTFY_PASSWORD")
				os.Unsetenv("NOTIFICATION_NTFY_AUTH_TOKEN")
			}
		})
	}

}

func setEnvCreds(setPassword, setToken, setAWS bool) {

	if setPassword {
		os.Setenv("NOTIFICATION_NTFY_PASSWORD", "1234")
	}

	if setToken {

		os.Setenv("NOTIFICATION_NTFY_AUTH_TOKEN", "abcd")
	}

	if setAWS {
		os.Setenv("AWS_ACCESS_KEY_ID", "1234")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "abcd")
		os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	}

}
