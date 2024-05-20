// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		id       int
		logLevel string
		opts     *slog.HandlerOptions
	}{
		{0, "DEBUG", &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: changeTimeFormat,
		}},
		{0, "INFO", &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			ReplaceAttr: changeTimeFormat,
		}},
		{0, "WARN", &slog.HandlerOptions{
			Level:       slog.LevelWarn,
			ReplaceAttr: changeTimeFormat,
		}},
		{0, "ERROR", &slog.HandlerOptions{
			Level:       slog.LevelError,
			ReplaceAttr: changeTimeFormat,
		}},
		{0, "UNKNOWN", &slog.HandlerOptions{}},
	}

	for index, test := range tests {
		test.id = index + 1
		results := logger(test.logLevel)
		ctx := context.Background()
		var lv slog.Level
		if test.opts.Level != nil {
			lv = test.opts.Level.Level()

		}
		got := results.Enabled(ctx, lv)
		if !got {
			t.Errorf("Test %d: Expected true, got %v", test.id, got)
		}

	}
}

func TestChangeTimeFormat(t *testing.T) {
	tests := []struct {
		id     int
		groups []string
		a      slog.Attr
	}{
		{0, []string{"time"}, slog.Attr{Key: slog.TimeKey, Value: slog.StringValue("2006/01/02 15:04:05")}},
		{0, []string{"time"}, slog.Attr{Key: slog.MessageKey, Value: slog.StringValue("2006/01/02 15:04:05")}},
	}

	for index, test := range tests {
		test.id = index + 1
		results := changeTimeFormat(test.groups, test.a)
		if results.Key != test.a.Key {
			t.Errorf("Test %d: Expected %v, got %v", test.id, test.a.Key, results.Key)
		}
		fmt.Println(results)
	}

}
