// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"log/slog"
	"runtime"
)

// LogError logs the error and the file, line, and function where the error occurred
func LogError(err error) {
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		slog.Debug("Error Information",
			slog.Group("File", slog.String("filename", file)),
			slog.Group("Line", slog.Int("line", line)),
			slog.Group("Function", slog.String("function", runtime.FuncForPC(pc).Name())),
		)
		slog.Debug("Error", "msg", err)
	} else {
		slog.Debug("Error", "msg", err)
	}
}
