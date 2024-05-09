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

	ok, token, err := verfyToken(cfg.Credentials.CredentialsFile)
	if err != nil {
		slog.Error("unable to verify token", "error", err)
		return err
	}

	if !ok {
		os.Exit(1)
	}

	data, err := user.GetUserProfileData(ctx, GlobalHTTPClient, token.AccessToken)
	if err != nil {
		return err
	}

	user.UserData = *data

	measurements, err := user.GetUserMeasurements(ctx, GlobalHTTPClient, token.AccessToken)
	if err != nil {
		return err
	}

	user.UserMesaurements = *measurements

	sleep, err := user.GetSleepCollection(ctx, GlobalHTTPClient, token.AccessToken, "")
	if err != nil {
		return err
	}

	user.SleepCollection = *sleep

	recovery, err := user.GetRecoveryCollection(ctx, GlobalHTTPClient, token.AccessToken, "")
	if err != nil {
		return err
	}

	user.RecoveryCollection = *recovery

	workout, err := user.GetWorkoutCollection(ctx, GlobalHTTPClient, token.AccessToken, "")
	if err != nil {
		internal.LogError(err)
		return err
	}

	user.WorkoutCollection = *workout

	finalDataRaw, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		internal.LogError(err)
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
			return err
		}
	default:
		err = fileExp.Export(finalDataRaw)
		if err != nil {
			return err
		}

	}

	return nil

}
