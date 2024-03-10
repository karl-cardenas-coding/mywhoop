package cmd

import (
	"context"
	"log/slog"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

// meCmd represents the me command
var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Review your Whoop profile and body measurements",
	Long:  "Review your Whoop profile and body measurements.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return me(rootCmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(meCmd)
}

func me(ctx context.Context) error {

	InitLogger()
	var user internal.User
	data, err := user.GetData(ctx, GlobalHTTPClient, AuthToken)
	if err != nil {
		return err
	}

	err = user.ExportData(data)
	if err != nil {
		slog.Error("unable to export data", err)
		return err
	}
	return nil

}
