// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
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
		output        string
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
			output:       "json",
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
			output:       "json",
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
			output:        "json",
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
			output:        "json",
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
			output:        "json",
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

			cFlags := cliFlags{
				dataLocation: test.dataLocation,
				output:       test.output,
			}

			if test.setEnvCreds {
				setEnvCreds(false, false, test.setAWScreds)
			}

			test.client = client

			exporterMethod, err := determineExporterExtension(test.cfg, test.client, cFlags)
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

			_ = os.Unsetenv("AWS_ACCESS_KEY_ID")
			_ = os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			_ = os.Unsetenv("AWS_DEFAULT_REGION")

		})
	}
}

func TestDetermineExporterExtension_SQLiteBranches(t *testing.T) {
	client := internal.CreateHTTPClient()

	t.Run("file+sqlite returns SQLiteExport", func(t *testing.T) {
		cfg := internal.ConfigurationData{
			Export: internal.ConfigExport{
				Method: "file",
				FileExport: export.FileExport{
					FilePath:       t.TempDir(),
					FileType:       "sqlite",
					FileName:       "user",
					FileNamePrefix: "test",
				},
			},
		}

		exporter, err := determineExporterExtension(cfg, client, cliFlags{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := exporter.(*internal.SQLiteExport); !ok {
			t.Fatalf("expected *internal.SQLiteExport, got %T", exporter)
		}
		if _, ok := exporter.(internal.UserExporter); !ok {
			t.Fatalf("expected exporter to satisfy internal.UserExporter")
		}
	})

	t.Run("file+sqlite via output flag override", func(t *testing.T) {
		// FileType in cfg is json, but the --output flag should override it
		// and route to the sqlite exporter.
		cfg := internal.ConfigurationData{
			Export: internal.ConfigExport{
				Method: "file",
				FileExport: export.FileExport{
					FilePath: t.TempDir(),
					FileType: "json",
					FileName: "user",
				},
			},
		}

		exporter, err := determineExporterExtension(cfg, client, cliFlags{output: "sqlite"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := exporter.(*internal.SQLiteExport); !ok {
			t.Fatalf("expected --output=sqlite to route to *internal.SQLiteExport, got %T", exporter)
		}
	})

	t.Run("s3+sqlite returns S3SQLiteExport", func(t *testing.T) {
		setEnvCreds(false, false, true)
		t.Cleanup(func() {
			_ = os.Unsetenv("AWS_ACCESS_KEY_ID")
			_ = os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			_ = os.Unsetenv("AWS_DEFAULT_REGION")
		})

		cfg := internal.ConfigurationData{
			Export: internal.ConfigExport{
				Method: "s3",
				AWSS3: export.AWS_S3{
					Region: "us-west-2",
					Bucket: "mybucket",
					FileConfig: export.FileExport{
						FilePath: t.TempDir(),
						FileType: "sqlite",
						FileName: "user",
					},
				},
			},
		}

		exporter, err := determineExporterExtension(cfg, client, cliFlags{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := exporter.(*internal.S3SQLiteExport); !ok {
			t.Fatalf("expected *internal.S3SQLiteExport, got %T", exporter)
		}
		if _, ok := exporter.(internal.UserExporter); !ok {
			t.Fatalf("expected s3+sqlite exporter to satisfy internal.UserExporter")
		}
	})

	t.Run("s3+json still returns AWS_S3", func(t *testing.T) {
		setEnvCreds(false, false, true)
		t.Cleanup(func() {
			_ = os.Unsetenv("AWS_ACCESS_KEY_ID")
			_ = os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			_ = os.Unsetenv("AWS_DEFAULT_REGION")
		})

		cfg := internal.ConfigurationData{
			Export: internal.ConfigExport{
				Method: "s3",
				AWSS3: export.AWS_S3{
					Region: "us-west-2",
					Bucket: "mybucket",
					FileConfig: export.FileExport{
						FileType: "json",
					},
				},
			},
		}

		exporter, err := determineExporterExtension(cfg, client, cliFlags{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := exporter.(*export.AWS_S3); !ok {
			t.Fatalf("expected *export.AWS_S3, got %T", exporter)
		}
	})
}

// recordingByteExporter only implements the byte-stream Export interface.
// writeUserToExporter should fall back to marshaling and call Export([]byte).
type recordingByteExporter struct {
	gotBytes []byte
	called   int
}

func (r *recordingByteExporter) Setup() error             { return nil }
func (r *recordingByteExporter) CleanUp() error           { return nil }
func (r *recordingByteExporter) Export(data []byte) error { r.gotBytes = data; r.called++; return nil }

// recordingUserExporter implements UserExporter so writeUserToExporter should
// dispatch directly without going through []byte.
type recordingUserExporter struct {
	gotUser     internal.User
	userCalled  int
	bytesCalled int
}

func (r *recordingUserExporter) Setup() error                  { return nil }
func (r *recordingUserExporter) CleanUp() error                { return nil }
func (r *recordingUserExporter) Export(_ []byte) error         { r.bytesCalled++; return nil }
func (r *recordingUserExporter) ExportUser(u internal.User) error {
	r.gotUser = u
	r.userCalled++
	return nil
}

// TestDetermineExporterExtension_RejectsInvalidFileType pins the defense-in-depth
// check in determineExporterExtension. PreRunE catches invalid --output values
// from the CLI, but the same FileType field also comes from the YAML config in
// server mode - this test makes sure both the file and s3 branches refuse to
// build an exporter when the type is something like "xlxs".
func TestDetermineExporterExtension_RejectsInvalidFileType(t *testing.T) {
	client := internal.CreateHTTPClient()

	t.Run("file method rejects xlxs", func(t *testing.T) {
		cfg := internal.ConfigurationData{
			Export: internal.ConfigExport{
				Method: "file",
				FileExport: export.FileExport{
					FilePath: t.TempDir(),
					FileType: "xlxs",
				},
			},
		}
		exp, err := determineExporterExtension(cfg, client, cliFlags{})
		if err == nil {
			t.Fatalf("expected error for file+xlxs, got exporter %T", exp)
		}
		if !strings.Contains(err.Error(), "unsupported fileType") {
			t.Errorf("error %q did not mention 'unsupported fileType'", err.Error())
		}
	})

	t.Run("file method rejects via --output override", func(t *testing.T) {
		cfg := internal.ConfigurationData{
			Export: internal.ConfigExport{
				Method: "file",
				FileExport: export.FileExport{
					FilePath: t.TempDir(),
					FileType: "json",
				},
			},
		}
		_, err := determineExporterExtension(cfg, client, cliFlags{output: "yaml"})
		if err == nil {
			t.Fatal("expected error when --output=yaml overrides a valid config")
		}
	})

	t.Run("s3 method rejects xlxs before building client", func(t *testing.T) {
		// Credentials intentionally unset: the validation should fire before
		// NewAwsS3Export is called, so we should see the fileType error rather
		// than a credential error.
		cfg := internal.ConfigurationData{
			Export: internal.ConfigExport{
				Method: "s3",
				AWSS3: export.AWS_S3{
					Region: "us-west-2",
					Bucket: "mybucket",
					FileConfig: export.FileExport{
						FileType: "xlxs",
					},
				},
			},
		}
		_, err := determineExporterExtension(cfg, client, cliFlags{})
		if err == nil {
			t.Fatal("expected error for s3+xlxs")
		}
		if !strings.Contains(err.Error(), "unsupported fileType") {
			t.Errorf("error %q did not mention 'unsupported fileType' (did validation run AFTER NewAwsS3Export?)", err.Error())
		}
	})
}

// TestMarshalUser_RejectsUnknown is a belt-and-suspenders test: upstream code
// guarantees marshalUser never sees an unsupported outputType, but we want a
// loud failure if anyone ever wires a new dispatch path that bypasses that
// validation. Previously the default branch silently returned JSON bytes.
func TestMarshalUser_RejectsUnknown(t *testing.T) {
	user := internal.User{UserData: internal.UserData{UserID: 1}}

	if _, err := marshalUser(user, "xlxs"); err == nil {
		t.Fatal("expected marshalUser(\"xlxs\") to return an error")
	}
	if _, err := marshalUser(user, "csv"); err == nil {
		t.Fatal("expected marshalUser(\"csv\") to return an error")
	}

	// sqlite is supported elsewhere but marshalUser is only for byte-stream
	// exporters; it is not expected to handle sqlite. Treat it as unsupported
	// at the marshalUser boundary.
	if _, err := marshalUser(user, "sqlite"); err == nil {
		t.Fatal("expected marshalUser(\"sqlite\") to return an error (sqlite is handled by UserExporter dispatch, not marshalUser)")
	}
}

func TestWriteUserToExporter_DispatchesToUserExporter(t *testing.T) {
	exp := &recordingUserExporter{}
	user := internal.User{UserData: internal.UserData{UserID: 99, Email: "a@b"}}

	if err := writeUserToExporter(user, exp, "json"); err != nil {
		t.Fatalf("writeUserToExporter: %v", err)
	}
	if exp.userCalled != 1 {
		t.Errorf("ExportUser call count = %d, want 1", exp.userCalled)
	}
	if exp.bytesCalled != 0 {
		t.Errorf("Export([]byte) was called %d times; should not be reached when UserExporter is satisfied", exp.bytesCalled)
	}
	if exp.gotUser.UserData.UserID != 99 {
		t.Errorf("ExportUser got UserID=%d, want 99", exp.gotUser.UserData.UserID)
	}
}

func TestWriteUserToExporter_FallsBackToBytes(t *testing.T) {
	exp := &recordingByteExporter{}
	user := internal.User{UserData: internal.UserData{UserID: 7, Email: "a@b"}}

	if err := writeUserToExporter(user, exp, "json"); err != nil {
		t.Fatalf("writeUserToExporter: %v", err)
	}
	if exp.called != 1 {
		t.Fatalf("Export([]byte) call count = %d, want 1", exp.called)
	}
	if len(exp.gotBytes) == 0 {
		t.Fatal("expected non-empty marshaled payload")
	}
	// Round-trip the bytes to confirm we got valid JSON of the user.
	var roundTripped internal.User
	if err := json.Unmarshal(exp.gotBytes, &roundTripped); err != nil {
		t.Fatalf("payload was not valid JSON: %v", err)
	}
	if roundTripped.UserData.UserID != 7 {
		t.Errorf("round-tripped UserID = %d, want 7", roundTripped.UserData.UserID)
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
				_ = os.Unsetenv("NOTIFICATION_NTFY_PASSWORD")
				_ = os.Unsetenv("NOTIFICATION_NTFY_AUTH_TOKEN")
			}
		})
	}

}

func setEnvCreds(setPassword, setToken, setAWS bool) {

	if setPassword {
		_ = os.Setenv("NOTIFICATION_NTFY_PASSWORD", "1234")
	}

	if setToken {

		_ = os.Setenv("NOTIFICATION_NTFY_AUTH_TOKEN", "abcd")
	}

	if setAWS {
		_ = os.Setenv("AWS_ACCESS_KEY_ID", "1234")
		_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "abcd")
		_ = os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	}

}

func TestGetFileType(t *testing.T) {

	tests := []struct {
		name     string
		cfg      internal.ConfigurationData
		expected string
	}{
		{
			name:     "File - json",
			expected: "json",
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					FileExport: export.FileExport{
						FileType: "json",
					},
				},
			},
		},
		{
			name:     "File - xlsx",
			expected: "xlsx",
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					FileExport: export.FileExport{
						FileType: "xlsx",
					},
				},
			},
		},
		{
			name:     "File - No Value Specified",
			expected: "json",
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{},
			},
		},
		{
			name:     "AWS S3 - xlsx",
			expected: "xlsx",
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					AWSS3: export.AWS_S3{
						FileConfig: export.FileExport{
							FileType: "xlsx",
						},
					},
				},
			},
		},
		{
			name:     "AWS S3 - json",
			expected: "json",
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					AWSS3: export.AWS_S3{
						FileConfig: export.FileExport{
							FileType: "json",
						},
					},
				},
			},
		},
		{
			name:     "AWS S3 - No Value Specified",
			expected: "json",
			cfg: internal.ConfigurationData{
				Export: internal.ConfigExport{
					AWSS3: export.AWS_S3{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			result := getFileType(test.cfg)
			if result != test.expected {
				t.Errorf("expected: %s, got: %s", test.expected, result)
			}
		})
	}
}
