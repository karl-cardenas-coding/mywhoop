// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/karl-cardenas-coding/mywhoop/notifications"
)

// determineExtension determines the notification extension to use and returns the appropriate notification.
func determineNotificationExtension(cfg internal.ConfigurationData) (internal.Notification, error) {

	var notificationMethod internal.Notification

	switch cfg.Notification.Method {
	case "ntfy":
		ntfy := notifications.NewNtfy()
		ntfy.ServerEndpoint = cfg.Notification.Ntfy.ServerEndpoint
		ntfy.SubscriptionID = cfg.Notification.Ntfy.SubscriptionID
		ntfy.UserName = cfg.Notification.Ntfy.UserName
		ntfy.Events = cfg.Notification.Ntfy.Events
		err := ntfy.SetUp()
		if err != nil {
			return notificationMethod, err
		}
		slog.Info("Ntfy notification method configured")
		notificationMethod = ntfy

	default:
		slog.Info("no notification method specified. Defaulting to stdout.")
		std := notifications.NewStdout()
		notificationMethod = std
	}

	return notificationMethod, nil

}

// determineExporterExtension determines the export extension to use and returns the appropriate export.
// When the resolved file type is sqlite, a SQLite-aware exporter is returned (SQLiteExport for file,
// S3SQLiteExport for s3). Otherwise the classic byte-stream exporter is returned.
func determineExporterExtension(cfg internal.ConfigurationData, client *http.Client, cFlags cliFlags) (internal.Export, error) {

	method := cfg.Export.Method
	if method == "" {
		slog.Info("no valid export method specified. Defaulting to file.")
		method = "file"
	}

	switch method {
	case "file":
		filePath := cfg.Export.FileExport.FilePath
		if cFlags.dataLocation != "" {
			filePath = cFlags.dataLocation
		}
		if cFlags.output != "" {
			cfg.Export.FileExport.FileType = cFlags.output
		}

		if cfg.Export.FileExport.FileType == "sqlite" {
			slog.Info("File + SQLite export method specified")
			return internal.NewSQLiteExport(
				filePath,
				cfg.Export.FileExport.FileName,
				cfg.Export.FileExport.FileNamePrefix,
			), nil
		}

		slog.Info("File export method specified")
		return export.NewFileExport(
			filePath,
			cfg.Export.FileExport.FileType,
			cfg.Export.FileExport.FileName,
			cfg.Export.FileExport.FileNamePrefix,
			cfg.Server.Enabled,
		), nil

	case "s3":
		if cFlags.dataLocation != "" {
			cfg.Export.AWSS3.FileConfig.FilePath = cFlags.dataLocation
		}
		if cFlags.output != "" {
			cfg.Export.AWSS3.FileConfig.FileType = cFlags.output
		}

		awsS3, err := export.NewAwsS3Export(
			cfg.Export.AWSS3.Region,
			cfg.Export.AWSS3.Bucket,
			cfg.Export.AWSS3.Profile,
			client,
			&cfg.Export.AWSS3.FileConfig,
			cfg.Server.Enabled,
		)
		if err != nil {
			return nil, errors.New("unable initialize AWS S3 export. Additional error context: " + err.Error())
		}

		if cfg.Export.AWSS3.FileConfig.FileType == "sqlite" {
			slog.Info("S3 + SQLite export method specified")
			sqlite := internal.NewSQLiteExport(
				cfg.Export.AWSS3.FileConfig.FilePath,
				cfg.Export.AWSS3.FileConfig.FileName,
				cfg.Export.AWSS3.FileConfig.FileNamePrefix,
			)
			return internal.NewS3SQLiteExport(sqlite, awsS3.S3Client, awsS3.Bucket), nil
		}

		slog.Info("AWS S3 export method specified")
		return awsS3, nil

	default:
		return nil, fmt.Errorf("unsupported export method %q", method)
	}
}

// getFileType determines the file type to use for the export based on the configuration and command line flags.
func getFileType(cfg internal.ConfigurationData) string {

	if cfg.Export.FileExport.FileType != "" {
		return cfg.Export.FileExport.FileType
	}

	if cfg.Export.AWSS3.FileConfig.FileType != "" {
		return cfg.Export.AWSS3.FileConfig.FileType
	}

	return "json"

}

// writeUserToExporter is the single code path that hands Whoop data to an exporter.
// Exporters that implement UserExporter (e.g. SQLite) consume the User struct directly.
// Everything else is serialized to bytes according to outputType and passed through
// the classic Export([]byte) path.
func writeUserToExporter(user internal.User, exporter internal.Export, outputType string) error {
	if ue, ok := exporter.(internal.UserExporter); ok {
		return ue.ExportUser(user)
	}

	data, err := marshalUser(user, outputType)
	if err != nil {
		return err
	}
	return exporter.Export(data)
}

// marshalUser converts a User into bytes appropriate for the given output format.
// Unknown formats fall back to pretty-printed JSON.
func marshalUser(user internal.User, outputType string) ([]byte, error) {
	switch strings.ToLower(outputType) {
	case "xlsx":
		return internal.ConvertToExcel(user)
	case "json", "":
		return json.MarshalIndent(user, "", "  ")
	default:
		return json.MarshalIndent(user, "", "  ")
	}
}
