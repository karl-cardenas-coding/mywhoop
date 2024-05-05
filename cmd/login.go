/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
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
	InitLogger()

	_, err := internal.GetToken()
	if err != nil {
		slog.Info("Error getting access token: %v", err)
		return err
	}
	return nil
}
