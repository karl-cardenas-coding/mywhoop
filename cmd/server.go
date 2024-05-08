// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// loginCmd represents the login command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server mode.",
	Long:  "Start myWhoop in server mode and download data from Whoop API on a regular basis.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return server(cmd.Context())
	},
}

var (
	// FirstRunDownload  downloads all data available from the Whoop API on the first run
	FirstRunDownload bool
)

func init() {
	serverCmd.PersistentFlags().BoolVar(&FirstRunDownload, "first-run-download", false, "Download all data available from the Whoop API on the first run.")
	rootCmd.AddCommand(serverCmd)
}

// EvaluateConfigOptions evaluates the configuration options for the server command
// Command line options take precedence over configuration file options.
func evaluateConfigOptions(firstRun bool, cfg *internal.ConfigurationData) error {

	if cfg.Export.Method == "" {
		slog.Info("No exporter specified. Defaulting to file.")
		cfg.Export.Method = "file"
	}

	if firstRun {
		slog.Info("First run download enabled")
		cfg.Server.FirstRunDownload = true
	}

	return nil

}

// login authenticates with Whoop API and gets an access token
func server(ctx context.Context) error {
	slog.Info("Server mode enabled")
	err := InitLogger()
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cfg := Configuration

	// Evaluate the configuration options
	err = evaluateConfigOptions(FirstRunDownload, &cfg)
	if err != nil {
		slog.Error("unable to evaluate configuration options", "error", err)
		return err
	}

	//Setup the exporters
	fileExp := export.FileExport{
		FilePath: Configuration.Export.FileExport.FilePath,
		FileType: Configuration.Export.FileExport.FileType,
		FileName: Configuration.Export.FileExport.FileName,
	}

	awsS3Exp := export.AWS_S3{
		Region: Configuration.Export.AWSS3.Region,
		Bucket: Configuration.Export.AWSS3.Bucket,
	}

	switch Configuration.Export.Method {
	case "file":

		err := fileExp.Setup()
		if err != nil {
			slog.Error("unable to setup file export", "error", err)
			return err
		}
	case "s3":
		err := awsS3Exp.Setup()
		if err != nil {
			slog.Error("unable to setup s3 export", "error", err)
			return err
		}

	default:
		slog.Error("unknown exporter", "exporter", cfg.Export.Method)
	}

	// Start the server entry point
	go func(c internal.ConfigurationData) {

		err := StartServer(ctx, c, GlobalHTTPClient)
		if err != nil {
			slog.Error("unable to start server", "error", err)
			os.Exit(1)
		}

	}(cfg)

	sig := <-sigs
	if sig == syscall.SIGINT || sig == syscall.SIGTERM {
		slog.Info("Server shutdown signal received")
		slog.Info("Cleaning up server resources")
		switch cfg.Export.Method {
		case "file":
			err := fileExp.CleanUp()
			if err != nil {
				slog.Error("unable to clean up file export", "error", err)
			}
		case "s3":
			err := awsS3Exp.CleanUp()
			if err != nil {
				slog.Error("unable to clean up s3 export", "error", err)
			}

		default:
			slog.Error("unknown exporter", "exporter", cfg.Export.Method)

		}

		slog.Info("Server shutdown complete")
		os.Exit(0)
	}

	return nil
}

// StartServer starts the long running server.
func StartServer(ctx context.Context, config internal.ConfigurationData, client *http.Client) error {

	ok, _, err := verfyToken(config.Credentials.CredentialsFile)
	if err != nil {
		slog.Error("unable to verify token", "error", err)
		return err
	}

	if !ok {
		os.Exit(1)
	}

	authTokenChannel := make(chan oauth2.Token)
	// This goroutine refreshes the token every minute.
	// The token is refreshed in the background so that the server can continue to run.
	go func() {
		ticker := time.NewTicker(55 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			slog.Info("Refreshing auth token token")
			currentToken, err := readTokenFromFile(config.Credentials.CredentialsFile)
			if err != nil {
				slog.Error("unable to read token file", "error", err)
				os.Exit(1)
			}

			token, err := internal.RefreshToken(ctx, currentToken.AccessToken, currentToken.RefreshToken, client)
			if err != nil {
				slog.Error("unable to refresh token", "error", err)
				os.Exit(1)
			}
			authTokenChannel <- token
		}
	}()

	// This goroutine writes the new token to a file.
	// This file is used when the Whoop API is called.
	go func() {

		for auth := range authTokenChannel {
			slog.Debug("New token generated:", auth.AccessToken[0:4], "....")

			data, err := json.MarshalIndent(auth, "", " ")
			if err != nil {
				slog.Error("unable to marshal token", "error", err)
				os.Exit(1)
			}

			err = os.WriteFile(config.Credentials.CredentialsFile, data, 0755)
			if err != nil {
				slog.Error("unable to write token file", "error", err)
				os.Exit(1)
			}
		}
	}()

	// This goroutine queries the Whoop API 24 hrs.
	go func() {

		ticker := time.NewTicker(24 * time.Hour)
		// ticker := time.NewTicker(1 * time.Minute) // DEBUG PURPOSES
		defer ticker.Stop()

		for range ticker.C {

			slog.Info("Starting data collection")

			token, err := readTokenFromFile(config.Credentials.CredentialsFile)
			if err != nil {
				slog.Error("unable to read token file", "error", err)
				os.Exit(1)
			}

			var user internal.User

			finalDataRaw, err := getData(ctx, user, client, token, &config.Server.FirstRunDownload)
			if err != nil {
				slog.Error("unable to get data", "error", err)
				os.Exit(1)
			}

			// Setup the exporters
			err = manageExporters(&config, finalDataRaw)
			if err != nil {
				slog.Error("unable to manage exporters", "error", err)
				os.Exit(1)
			}

			slog.Info("Data collection complete")
		}

	}()

	return nil
}

// manageExporters manages the exporters based on the configuration received
func manageExporters(cfg *internal.ConfigurationData, data []byte) error {

	if cfg.Export.FileExport.FileNamePrefix == "" {
		cfg.Export.FileExport.FileNamePrefix = "user"
	}
	// Configure the filename to ensure uniqueness
	fileName := fmt.Sprintf("%s_%s", cfg.Export.FileExport.FileNamePrefix, internal.GetCurrentDate())

	fileExp := export.FileExport{
		FilePath: cfg.Export.FileExport.FilePath,
		FileType: cfg.Export.FileExport.FileType,
		FileName: fileName,
	}

	awsS3Exp := export.AWS_S3{
		Region: cfg.Export.AWSS3.Region,
		Bucket: cfg.Export.AWSS3.Bucket,
	}

	switch cfg.Export.Method {
	case "file":
		err := fileExp.Export(data)
		if err != nil {
			slog.Error("unable to export data with the file exporter", "error", err)
			internal.LogError(err)
			return err

		}

	case "s3":
		err := awsS3Exp.Export(data)
		if err != nil {
			slog.Error("unable to export data with the s3 exporter", "error", err)
			internal.LogError(err)
			return err
		}
	default:
		slog.Error("unknown exporter", "exporter", cfg.Export.Method)

	}

	return nil

}

// getData queries the Whoop API and gets the user data
func getData(ctx context.Context, user internal.User, client *http.Client, token oauth2.Token, firstDownload *bool) ([]byte, error) {

	if firstDownload == nil {
		slog.Debug("firstDownload is nil. Unable to determine if this is the first download")
		firstDownload = new(bool)
		*firstDownload = false
	}

	if !*firstDownload {
		startTime, endTime := internal.GenerateLast24HoursString()
		filterString := fmt.Sprintf("start=%s&end=%s", startTime, endTime)

		slog.Debug("Filter string", "filter", filterString)

		sleep, err := user.GetSleepCollection(ctx, client, token.AccessToken, filterString)
		if err != nil {
			internal.LogError(err)
		}

		user.SleepCollection = *sleep

		recovery, err := user.GetRecoveryCollection(ctx, client, token.AccessToken, filterString)
		if err != nil {
			internal.LogError(err)
		}

		user.RecoveryCollection = *recovery

		workout, err := user.GetWorkoutCollection(ctx, client, token.AccessToken, filterString)
		if err != nil {
			internal.LogError(err)
		}

		user.WorkoutCollection = *workout
	}

	if *firstDownload {

		data, err := user.GetUserProfileData(ctx, client, token.AccessToken)
		if err != nil {
			internal.LogError(err)
		}

		user.UserData = *data

		measurements, err := user.GetUserMeasurements(ctx, client, token.AccessToken)
		if err != nil {
			internal.LogError(err)
		}

		user.UserMesaurements = *measurements

		sleep, err := user.GetSleepCollection(ctx, client, token.AccessToken, "")
		if err != nil {
			internal.LogError(err)
		}

		user.SleepCollection = *sleep

		recovery, err := user.GetRecoveryCollection(ctx, client, token.AccessToken, "")
		if err != nil {
			internal.LogError(err)
		}

		user.RecoveryCollection = *recovery

		workout, err := user.GetWorkoutCollection(ctx, client, token.AccessToken, "")
		if err != nil {
			internal.LogError(err)
		}

		user.WorkoutCollection = *workout
		// Set to false so that the entire data is not downloaded again
		*firstDownload = false
	}

	finalDataRaw, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		internal.LogError(err)
	}

	return finalDataRaw, nil

}

func verfyToken(filePath string) (bool, oauth2.Token, error) {

	// verify the file exists
	_, err := os.Stat(filePath)
	if err != nil {
		slog.Error("Token file does not exist", "error", err)
		return false, oauth2.Token{}, err
	}

	token, err := readTokenFromFile(filePath)
	if err != nil {
		slog.Error("unable to read token file", "error", err)
		return false, oauth2.Token{}, err
	}

	if !token.Valid() {
		slog.Error("Auth token is not valid")
		return false, oauth2.Token{}, nil
	}

	slog.Info("Auth token is valid")

	return true, token, nil
}

// readTokenFromFile reads a token from a file and returns it as an oauth2.Token
func readTokenFromFile(filePath string) (oauth2.Token, error) {

	f, err := os.Open(filePath)
	if err != nil {
		slog.Error("unable to open token file", "error", err)
		return oauth2.Token{}, err
	}
	defer f.Close()

	var token oauth2.Token
	json.NewDecoder(f).Decode(&token)

	return token, nil
}
