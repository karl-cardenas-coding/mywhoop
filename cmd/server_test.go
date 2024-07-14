// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
)

func TestEvaluateConfigOptions(t *testing.T) {

	dt := &internal.ConfigurationData{
		Export: internal.ConfigExport{
			Method: "file",
			FileExport: export.FileExport{
				FilePath: "tests/data/",
			},
		},
		Server: internal.Server{
			Enabled: true,
		},
	}

	expectedExporter := "file"

	err := evaluateConfigOptions(dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Second test
	dt.Export.Method = ""

	err = evaluateConfigOptions(dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Third test

	dt = &internal.ConfigurationData{}
	err = evaluateConfigOptions(dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Fourth test
	dt = &internal.ConfigurationData{
		Server: internal.Server{
			Enabled: true,
		},
	}
	err = evaluateConfigOptions(dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

}

func TestLoggerConverter(t *testing.T) {

	tests := []struct {
		name     string
		level    string
		expected gocron.LogLevel
	}{
		{
			name:     "Test Debug",
			level:    "debug",
			expected: gocron.LogLevelDebug,
		},
		{
			name:     "Test Info",
			level:    "info",
			expected: gocron.LogLevelInfo,
		},
		{
			name:     "Test Warn",
			level:    "warn",
			expected: gocron.LogLevelWarn,
		},
		{
			name:     "Test Error",
			level:    "error",
			expected: gocron.LogLevelError,
		},
		{
			name:     "Test Non Allowed Value",
			level:    "fatal",
			expected: gocron.LogLevel(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loggerConverter(tt.level)
			if result != tt.expected {
				t.Errorf("Expected %v, got: %v", tt.expected, result)
			}
		})
	}

}

func TestJwtRefreshDurationValidator(t *testing.T) {

	tests := []struct {
		name     string
		duration int
		expected time.Duration
	}{
		{
			name:     "Test Zero Duration",
			duration: 0,
			expected: internal.DEFAULT_SERVER_TOKEN_REFRESH_CRON_SCHEDULE,
		},
		{
			name:     "Test Negative Duration",
			duration: -1,
			expected: internal.DEFAULT_SERVER_TOKEN_REFRESH_CRON_SCHEDULE,
		},
		{
			name:     "Test Positive Duration - 5 min",
			duration: 5,
			expected: 5 * time.Minute,
		},
		{
			name:     "Test Positive Duration - 55 min",
			duration: 55,
			expected: 55 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jwtRefreshDurationValidator(tt.duration)
			if result != tt.expected {
				t.Errorf("Expected %v, got: %v", tt.expected, result)
			}
		})
	}

}
