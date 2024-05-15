// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
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
	Short: "Dump all your Whoop data to a file.",
	Long:  "Dump all your Whoop data to a file.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return dump(rootCmd.Context())

	},
}

func init() {
	rootCmd.AddCommand(meCmd)
}

func dump(ctx context.Context) error {

	var user internal.User

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

	ntfy := &notifications.Ntfy{
		ServerEndpoint: cfg.Notification.Ntfy.ServerEndpoint,
		SubscriptionID: cfg.Notification.Ntfy.SubscriptionID,
		UserName:       cfg.Notification.Ntfy.UserName,
	}
	var notificationMethod notifications.Notification

	switch Configuration.Notification.Method {
	case "ntfy":
		err = ntfy.SetUp()
		if err != nil {
			return err
		}
		notificationMethod = ntfy
	default:
		slog.Info("no notification method specified. Defaulting to stdout.")
	}

	data, err := user.GetUserProfileData(ctx, GlobalHTTPClient, token.AccessToken)
	if err != nil {
		internal.LogError(err)
		externalNotificaton(notificationMethod, []byte(err.Error()))
		return err
	}

	user.UserData = *data

	measurements, err := user.GetUserMeasurements(ctx, GlobalHTTPClient, token.AccessToken)
	if err != nil {
		internal.LogError(err)
		externalNotificaton(notificationMethod, []byte(err.Error()))
		return err
	}

	user.UserMesaurements = *measurements

	sleep, err := user.GetSleepCollection(ctx, GlobalHTTPClient, token.AccessToken, "")
	if err != nil {
		internal.LogError(err)
		externalNotificaton(notificationMethod, []byte(err.Error()))
		return err
	}

	user.SleepCollection = *sleep

	recovery, err := user.GetRecoveryCollection(ctx, GlobalHTTPClient, token.AccessToken, "")
	if err != nil {
		internal.LogError(err)
		externalNotificaton(notificationMethod, []byte(err.Error()))
		return err
	}

	user.RecoveryCollection = *recovery

	workout, err := user.GetWorkoutCollection(ctx, GlobalHTTPClient, token.AccessToken, "")
	if err != nil {
		internal.LogError(err)
		externalNotificaton(notificationMethod, []byte(err.Error()))
		return err
	}

	user.WorkoutCollection = *workout

	cycle, err := user.GetCycleCollection(ctx, GlobalHTTPClient, token.AccessToken, "")
	if err != nil {
		internal.LogError(err)
		externalNotificaton(notificationMethod, []byte(err.Error()))
		return err
	}

	user.CycleCollection = *cycle

	finalDataRaw, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		internal.LogError(err)
		externalNotificaton(notificationMethod, []byte(err.Error()))
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
			externalNotificaton(notificationMethod, []byte(err.Error()))
			return err
		}
	default:
		err = fileExp.Export(finalDataRaw)
		if err != nil {
			externalNotificaton(notificationMethod, []byte(err.Error()))
			return err
		}

	}

	return nil

}

// externalNotificaton sends a notification to the user using the specified notification method.
func externalNotificaton(notificationMethod notifications.Notification, msg []byte) {

	if notificationMethod == nil {
		// no notification method specified, do nothing
		return
	}

	err := notificationMethod.Send([]byte("unable to fetch data"))
	if err != nil {
		internal.LogError(err)
		slog.Info("unable to send external notification", "error", err)
	}
}
