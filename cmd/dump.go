package cmd

import (
	"context"
	"log/slog"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

// meCmd represents the me command
var meCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all your data to a file.",
	Long:  "Dump all your data to a file.",
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

	data, err := user.GetData(ctx, GlobalHTTPClient, AuthToken)
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

	err = user.ExportDataToFile("")
	if err != nil {
		slog.Error("unable to export data", err)
		return err
	}
	return nil

}
