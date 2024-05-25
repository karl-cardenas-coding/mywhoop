// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
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

	var notificationMethod notifications.Notification

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
		fmt.Println("Events", ntfy.Events)
	default:
		slog.Info("no notification method specified. Defaulting to stdout.")
	}

	data, err := user.GetUserProfileData(ctx, client, internal.DEFAULT_WHOOP_API_USER_DATA_URL, token.AccessToken, ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	user.UserData = *data

	measurements, err := user.GetUserMeasurements(ctx, client, internal.DEFAULT_WHOOP_API_USER_MEASUREMENT_DATA_URL, token.AccessToken, ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	user.UserMesaurements = *measurements

	sleep, err := user.GetSleepCollection(ctx, client, internal.DEFAULT_WHOOP_API_USER_SLEEP_DATA_URL, token.AccessToken, "", ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	user.SleepCollection = *sleep

	recovery, err := user.GetRecoveryCollection(ctx, client, internal.DEFAULT_WHOOP_API_RECOVERY_DATA_URL, token.AccessToken, "", ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	user.RecoveryCollection = *recovery

	workout, err := user.GetWorkoutCollection(ctx, client, internal.DEFAULT_WHOOP_API_WORKOUT_DATA_URL, token.AccessToken, "", ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	user.WorkoutCollection = *workout

	cycle, err := user.GetCycleCollection(ctx, client, internal.DEFAULT_WHOOP_API_CYCLE_DATA_URL, token.AccessToken, "", ua)
	if err != nil {
		internal.LogError(err)
		notifyErr := notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	user.CycleCollection = *cycle

	finalDataRaw, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		internal.LogError(err)
		notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
		return err
	}

	fileExp := export.FileExport{
		FilePath:       Configuration.Export.FileExport.FilePath,
		FileType:       Configuration.Export.FileExport.FileType,
		FileName:       Configuration.Export.FileExport.FileName,
		FileNamePrefix: Configuration.Export.FileExport.FileNamePrefix,
	}

	switch Configuration.Export.Method {
	case "file":
		err = fileExp.Export(finalDataRaw)
		if err != nil {
			notifyErr := notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
			return err
		}
		slog.Info("Data exported successfully", "file", fileExp.FileName)
	default:
		err = fileExp.Export(finalDataRaw)
		if err != nil {
			notifyErr := notifications.Publish(client, notificationMethod, []byte(err.Error()), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
			return err
		}

	}
	slog.Info("All Whoop data downloaded successfully")
	notifyErr := notifications.Publish(client, notificationMethod, []byte("Successfully downloaded all Whoop data."), internal.EventSuccess.String())
	if notifyErr != nil {
		slog.Error("unable to send notification", "error", notifyErr)
	}

	return nil

}
