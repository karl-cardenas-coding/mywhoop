// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/karl-cardenas-coding/mywhoop/notifications"
	"github.com/spf13/cobra"
)

// meCmd represents the me command
var meCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all your Whoop data to a file or another form of export.",
	Long:  "Dump all your Whoop data to a file or another form of export.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return dump(rootCmd.Context())

	},
}

func init() {
	rootCmd.AddCommand(meCmd)
}

func dump(ctx context.Context) error {

	var user internal.User
	var ua string = UserAgent

	client := internal.CreateHTTPClient()

	err := InitLogger(&Configuration)
	if err != nil {
		return err
	}

	cfg := Configuration

	ok, token, err := internal.VerfyToken(cfg.Credentials.CredentialsFile)
	if err != nil {
		slog.Error("unable to verify token", "error", err)
		return err
	}

	if !ok {
		os.Exit(1)
	}

	var notificationMethod internal.Notification

	switch cfg.Notification.Method {
	case "ntfy":
		ntfy := notifications.NewNtfy()
		ntfy.ServerEndpoint = cfg.Notification.Ntfy.ServerEndpoint
		ntfy.SubscriptionID = cfg.Notification.Ntfy.SubscriptionID
		ntfy.UserName = cfg.Notification.Ntfy.UserName
		ntfy.Events = cfg.Notification.Ntfy.Events
		err = ntfy.SetUp()
		if err != nil {
			return err
		}
		notificationMethod = ntfy
		slog.Info("Ntfy notification method configured")
	default:
		slog.Info("no notification method specified. Defaulting to stdout.")
		std := notifications.NewStdout()
		notificationMethod = std
	}

	data, err := user.GetUserProfileData(ctx, client, internal.DEFAULT_WHOOP_API_USER_DATA_URL, token.AccessToken, ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	user.UserData = *data

	measurements, err := user.GetUserMeasurements(ctx, client, internal.DEFAULT_WHOOP_API_USER_MEASUREMENT_DATA_URL, token.AccessToken, ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	user.UserMesaurements = *measurements

	sleep, err := user.GetSleepCollection(ctx, client, internal.DEFAULT_WHOOP_API_USER_SLEEP_DATA_URL, token.AccessToken, "", ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	sleep.NextToken = ""
	user.SleepCollection = *sleep

	recovery, err := user.GetRecoveryCollection(ctx, client, internal.DEFAULT_WHOOP_API_RECOVERY_DATA_URL, token.AccessToken, "", ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	recovery.NextToken = ""
	user.RecoveryCollection = *recovery

	workout, err := user.GetWorkoutCollection(ctx, client, internal.DEFAULT_WHOOP_API_WORKOUT_DATA_URL, token.AccessToken, "", ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	workout.NextToken = ""
	user.WorkoutCollection = *workout

	cycle, err := user.GetCycleCollection(ctx, client, internal.DEFAULT_WHOOP_API_CYCLE_DATA_URL, token.AccessToken, "", ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}
	cycle.NextToken = ""
	user.CycleCollection = *cycle

	finalDataRaw, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		internal.LogError(err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	switch cfg.Export.Method {
	case "file":
		fileExp := export.NewFileExport(Configuration.Export.FileExport.FilePath,
			Configuration.Export.FileExport.FileType,
			Configuration.Export.FileExport.FileName,
			Configuration.Export.FileExport.FileNamePrefix,
			false,
		)
		err = fileExp.Export(finalDataRaw)
		if err != nil {
			notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
			return err
		}
		slog.Info("Data exported successfully", "file", fileExp.FileName)
	case "s3":
		awsS3, err := export.NewAwsS3Export(cfg.Export.AWSS3.Region,
			cfg.Export.AWSS3.Bucket,
			cfg.Export.AWSS3.Profile,
			client,
			&cfg.Export.AWSS3.FileConfig,
			false,
		)
		if err != nil {
			return errors.New("unable initialize AWS S3 export. Additional error context: " + err.Error())
		}
		err = awsS3.Export(finalDataRaw)
		if err != nil {
			notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
			return errors.New("unable to export data to AWS S3. Additional error context: " + err.Error())
		}

	default:
		slog.Info("no export method specified. Defaulting to file.")
		fileExp := export.NewFileExport(Configuration.Export.FileExport.FilePath,
			Configuration.Export.FileExport.FileType,
			Configuration.Export.FileExport.FileName,
			Configuration.Export.FileExport.FileNamePrefix,
			false,
		)
		err = fileExp.Export(finalDataRaw)
		if err != nil {
			notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
			return err
		}

	}
	slog.Info("All Whoop data downloaded successfully")
	if notificationMethod != nil {
		err = notificationMethod.Publish(client, []byte("Successfully downloaded all Whoop data."), internal.EventSuccess.String())
		if err != nil {
			slog.Error("unable to send notification", "error", err)
		}
	}

	return nil

}
