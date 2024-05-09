// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log/slog"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Whoop API and get an access token",
	Long:  "Authenticate with Whoop API and get an access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		return login()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

// login authenticates with Whoop API and gets an access token
func login() error {
	err := InitLogger(&Configuration)
	if err != nil {
		return err
	}

	_, err = internal.GetToken("token.json")
	if err != nil {
		slog.Info("Error getting access token: %v", err)
		return err
	}
	return nil
}
