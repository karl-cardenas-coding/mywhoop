// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

var (
	dataLocation string
	filter       string
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all your Whoop data to a file or another form of export.",
	Long:  "Dump all your Whoop data to a file or another form of export.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return dump(rootCmd.Context())

	},
}

func init() {
	dumpCmd.PersistentFlags().StringVarP(&dataLocation, "location", "l", "", "The location to dump the data to. Default is the current directory's data/ folder.")
	dumpCmd.PersistentFlags().StringVarP(&filter, "filter", "f", "", "Provide a filter string to narrow down the data to download. For example, start=2024-01-01T00:00:00.000Z&end=2022-04-01T00:00:00.000Z")

	rootCmd.AddCommand(dumpCmd)
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
	cfg.Server.Enabled = false

	ok, token, err := internal.VerfyToken(cfg.Credentials.CredentialsFile)
	if err != nil {
		slog.Error("unable to verify token", "error", err)
		return err
	}

	if !ok {
		os.Exit(1)
	}

	notificationMethod, err := determineNotificationExtension(cfg)
	if err != nil {
		return err
	}

	if filter != "" {
		slog.Info("Filtering data with:", "filter", filter)
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

	sleep, err := user.GetSleepCollection(ctx, client, internal.DEFAULT_WHOOP_API_USER_SLEEP_DATA_URL, token.AccessToken, filter, ua)
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

	recovery, err := user.GetRecoveryCollection(ctx, client, internal.DEFAULT_WHOOP_API_RECOVERY_DATA_URL, token.AccessToken, filter, ua)
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

	workout, err := user.GetWorkoutCollection(ctx, client, internal.DEFAULT_WHOOP_API_WORKOUT_DATA_URL, token.AccessToken, filter, ua)
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

	cycle, err := user.GetCycleCollection(ctx, client, internal.DEFAULT_WHOOP_API_CYCLE_DATA_URL, token.AccessToken, filter, ua)
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

	exporterMethod, err := determineExporterExtension(cfg, client)
	if err != nil {
		slog.Error("unable to determine export method", "error", err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	err = exporterMethod.Export(finalDataRaw)
	if err != nil {
		slog.Error("unable to export data", "error", err)
		notifyErr := notificationMethod.Publish(client, []byte(err.Error()), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	slog.Info("All Whoop data downloaded and exported successfully")
	if notificationMethod != nil {
		err = notificationMethod.Publish(client, []byte("Successfully downloaded all Whoop data."), internal.EventSuccess.String())
		if err != nil {
			slog.Error("unable to send notification", "error", err)
		}
	}

	return nil

}
