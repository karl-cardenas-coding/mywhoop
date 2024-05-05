package cmd

import (
	"context"
	"encoding/json"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

// meCmd represents the me command
var meCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all your Whoop data to a file.",
	Long:  "Dump all your Whoo[ data to a file.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return dump(rootCmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(meCmd)
}

func dump(ctx context.Context) error {

	var user internal.User

	InitLogger()

	// authToken, err := internal.RefreshToken(ctx, AuthToken, RefreshToken, GlobalHTTPClient)
	// if err != nil {
	// 	internal.LogError(err)
	// 	return err
	// }
	// slog.Debug(authToken)
	// slog.Info("Token refreshed")

	data, err := user.GetUserProfileData(ctx, GlobalHTTPClient, AuthToken)
	if err != nil {
		return err
	}

	user.UserData = *data

	measurements, err := user.GetUserMeasurements(ctx, GlobalHTTPClient, AuthToken)
	if err != nil {
		return err
	}

	user.UserMesaurements = *measurements

	sleep, err := user.GetSleepCollection(ctx, GlobalHTTPClient, AuthToken, "")
	if err != nil {
		return err
	}

	user.SleepCollection = *sleep

	recovery, err := user.GetRecoveryCollection(ctx, GlobalHTTPClient, AuthToken, "")
	if err != nil {
		return err
	}

	user.RecoveryCollection = *recovery

	workout, err := user.GetWorkoutCollection(ctx, GlobalHTTPClient, AuthToken, "")
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
		FilePath: Configuration.Export.FileExport.FilePath,
		FileType: Configuration.Export.FileExport.FileType,
		FileName: Configuration.Export.FileExport.FileName,
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
