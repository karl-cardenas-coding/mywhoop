// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"log/slog"
	"net/http"

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
// The parameter isServerMode is used to determine if the exporter is being used in server mode. Use this flag to set server mode defaults.
func determineExporterExtension(cfg internal.ConfigurationData, client *http.Client, cFlags cliFlags) (internal.Export, error) {

	var (
		filePath string
		exporter internal.Export
	)

	switch cfg.Export.Method {
	case "file":
		if cFlags.dataLocation == "" {
			filePath = cfg.Export.FileExport.FilePath
		}

		if cFlags.dataLocation != "" {
			filePath = cFlags.dataLocation
		}

		if cFlags.output != "" {
			cfg.Export.FileExport.FileType = cFlags.output
		}

		fileExp := export.NewFileExport(filePath,
			cfg.Export.FileExport.FileType,
			cfg.Export.FileExport.FileName,
			cfg.Export.FileExport.FileNamePrefix,
			cfg.Server.Enabled,
		)
		slog.Info("File export method specified")
		exporter = fileExp

	case "s3":
		slog.Info("AWS S3 export method specified")
		if cFlags.dataLocation != "" {
			cfg.Export.AWSS3.FileConfig.FilePath = cFlags.dataLocation
		}

		if cFlags.output != "" {
			cfg.Export.AWSS3.FileConfig.FileType = cFlags.output
		}

		awsS3, err := export.NewAwsS3Export(cfg.Export.AWSS3.Region,
			cfg.Export.AWSS3.Bucket,
			cfg.Export.AWSS3.Profile,
			client,
			&cfg.Export.AWSS3.FileConfig,
			cfg.Server.Enabled,
		)
		if err != nil {
			return exporter, errors.New("unable initialize AWS S3 export. Additional error context: " + err.Error())
		}
		exporter = awsS3

	default:
		if cFlags.dataLocation == "" {
			filePath = cfg.Export.FileExport.FilePath
		} else {
			filePath = cFlags.dataLocation
		}

		if cFlags.output != "" {
			cfg.Export.FileExport.FileType = cFlags.output
		}

		slog.Info("no valid export method specified. Defaulting to file.")

		fileExp := export.NewFileExport(filePath,
			cfg.Export.FileExport.FileType,
			cfg.Export.FileExport.FileName,
			cfg.Export.FileExport.FileNamePrefix,
			cfg.Server.Enabled,
		)
		exporter = fileExp

	}

	return exporter, nil

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
