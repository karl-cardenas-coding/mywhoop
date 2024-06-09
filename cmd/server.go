// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/karl-cardenas-coding/mywhoop/notifications"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

// loginCmd represents the login command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server mode.",
	Long:  "Start MyWhoop in server mode and download data from Whoop API on a regular basis.",
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

	err := InitLogger(&Configuration)
	if err != nil {
		return err
	}
	cfg := Configuration
	client := internal.CreateHTTPClient()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var ua string = UserAgent

	// Evaluate the configuration options
	err = evaluateConfigOptions(FirstRunDownload, &cfg)
	if err != nil {
		slog.Error("unable to evaluate configuration options", "error", err)
		return err
	}
	var exportSelected internal.Export
	// Initialize the data exporters
	switch cfg.Export.Method {
	case "file":
		fileExp := export.NewFileExport(cfg.Export.FileExport.FilePath,
			cfg.Export.FileExport.FileType,
			cfg.Export.FileExport.FileName,
			cfg.Export.FileExport.FileNamePrefix,
			true,
		)

		if cfg.Export.FileExport.FileNamePrefix == "" {
			cfg.Export.FileExport.FileNamePrefix = "user"
		}

		exportSelected = fileExp
	case "s3":
		awsS3Exp, err := export.NewAwsS3Export(cfg.Export.AWSS3.Region, cfg.Export.AWSS3.Bucket, cfg.Export.AWSS3.Profile, client, &cfg.Export.AWSS3.FileConfig, true)
		if err != nil {
			slog.Error("unable to initialize AWS S3 export", "error", err)
			return err
		}

		exportSelected = awsS3Exp
		slog.Info("AWS S3 export method specified")
	default:
		slog.Error("unknown exporter", "exporter", cfg.Export.Method)
	}

	// Setup the notification method
	err = exportSelected.Setup()
	if err != nil {
		slog.Error("unable to setup data exporter", "error", err)
		return err
	}

	var notificationMethod internal.Notification

	// Initialize the notification method
	switch Configuration.Notification.Method {
	case "ntfy":
		ntfy := notifications.NewNtfy()
		ntfy.ServerEndpoint = cfg.Notification.Ntfy.ServerEndpoint
		ntfy.SubscriptionID = cfg.Notification.Ntfy.SubscriptionID
		ntfy.UserName = cfg.Notification.Ntfy.UserName
		ntfy.Events = cfg.Notification.Ntfy.Events
		slog.Info("Ntfy notification method specified")
		notificationMethod = ntfy
	default:
		slog.Info("No notification method specified. Defaulting to stdout.")
		std := notifications.NewStdout()
		notificationMethod = std

	}

	// Setup the notification method
	if notificationMethod != nil {
		err = notificationMethod.SetUp()
		if err != nil {
			slog.Error("unable to setup notification method", "error", err)
			return err
		}
	}

	g, ctx := errgroup.WithContext(ctx)

	// Download the latest data for the past 24 hrs and if FirstRunDownload is enabled, all of the data.
	g.Go(func() error {

		ok, _, err := internal.VerfyToken(cfg.Credentials.CredentialsFile)
		if err != nil {
			slog.Error("unable to verify authentication token", "error", err)
			return err
		}

		if !ok {
			return errors.New("auth token is invalid or expired")
		}

		slog.Info("Starting data collection")

		token, err := internal.ReadTokenFromFile(cfg.Credentials.CredentialsFile)
		if err != nil {
			slog.Error("unable to read token file", "error", err)
			return err
		}

		var user internal.User

		finalDataRaw, err := getData(ctx, user, client, token, &cfg.Server.FirstRunDownload, ua)
		if err != nil {
			slog.Error("unable to get data", "error", err)
			return err
		}

		err = exportSelected.Export(finalDataRaw)
		if err != nil {
			slog.Error("unable to export data", "error", err)
			return err
		}

		err = exportSelected.CleanUp()
		if err != nil {
			slog.Error("unable to clean up export", "error", err)
			return err
		}

		slog.Info("Data collection complete")
		err = notificationMethod.Publish(client, []byte("Initial data collection complete."), internal.EventSuccess.String())
		if err != nil {
			slog.Error("unable to send notification", "error", err)
		}

		return nil

	})
	// Handle a sigterm if the cron logic has not started yet
	// firstSigOp := <-sigs
	// if firstSigOp == syscall.SIGINT || firstSigOp == syscall.SIGTERM {
	// 	slog.Info("program interrupt received")
	// 	os.Exit(0)
	// }
	if err := g.Wait(); err != nil {
		notifyErr := notificationMethod.Publish(client, []byte(fmt.Sprintf("An error occured during the initial data collection. Additional error message: \n %s", err)), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}
	// Start the server entry point
	go func(c internal.ConfigurationData) {

		err := StartServer(ctx, c, client, exportSelected, notificationMethod)
		if err != nil {
			slog.Error("unable to start server", "error", err)
			notifyErr := notificationMethod.Publish(client, []byte(fmt.Sprintf("unable to start server. Additional error message: \n %s", err)), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
			os.Exit(1)
		}

	}(cfg)

	sig := <-sigs
	if sig == syscall.SIGINT || sig == syscall.SIGTERM {
		slog.Info("Server shutdown signal received")
		slog.Info("Cleaning up server resources")
		err := exportSelected.CleanUp()
		if err != nil {
			slog.Error("unable to clean up export", "error", err)
			notifyErr := notificationMethod.Publish(client, []byte(fmt.Sprintf("unable to clean up export. Additional error message: \n %s", err)), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
		}
		slog.Info("Server shutdown complete")
		os.Exit(0)
	}

	return nil
}

// StartServer starts the long running server.
func StartServer(ctx context.Context, config internal.ConfigurationData, client *http.Client, exp internal.Export, notify internal.Notification) error {

	ok, _, err := internal.VerfyToken(config.Credentials.CredentialsFile)
	if err != nil {
		slog.Error("unable to verify token", "error", err)
		notifyErr := notify.Publish(client, []byte("Unable to verify the existing token during the token refresh process."), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		return err
	}

	if !ok {
		slog.Error("auth token is invalid or expired")
		notifyErr := notify.Publish(client, []byte("The authentication token is invalid or expired."), internal.EventErrors.String())
		if notifyErr != nil {
			slog.Error("unable to send notification", "error", notifyErr)
		}
		os.Exit(1)
	}

	authTokenChannel := make(chan oauth2.Token)
	// This goroutine refreshes the token every minute.
	// The token is refreshed in the background so that the server can continue to run.
	go func() {
		ticker := time.NewTicker(55 * time.Minute)
		// ticker := time.NewTicker(2 * time.Minute) // DEBUG PURPOSES
		defer ticker.Stop()

		for range ticker.C {
			slog.Info("Refreshing auth token token")
			currentToken, err := internal.ReadTokenFromFile(config.Credentials.CredentialsFile)
			if err != nil {
				slog.Error("unable to read token file", "error", err)
				notifyErr := notify.Publish(client, []byte(fmt.Sprintf("Unable to read the authentication token from file. Additional context below: \n %s", err)), internal.EventErrors.String())
				if notifyErr != nil {
					slog.Error("unable to send notification", "error", notifyErr)
				}
				os.Exit(1)
			}

			auth := internal.AuthRequest{
				AuthToken:        currentToken.AccessToken,
				RefreshToken:     currentToken.RefreshToken,
				Client:           client,
				ClientID:         os.Getenv("WHOOP_CLIENT_ID"),
				ClientSecret:     os.Getenv("WHOOP_CLIENT_SECRET"),
				TokenURL:         internal.DEFAULT_ACCESS_TOKEN_URL,
				AuthorizationURL: internal.DEFAULT_AUTHENTICATION_URL,
			}

			token, err := internal.RefreshToken(ctx, auth)
			if err != nil {
				slog.Error("unable to refresh token", "error", err)
				notifyErr := notify.Publish(client, []byte(fmt.Sprintf("Unable to refresh the authentication token. Additional context below: \n %s", err)), internal.EventErrors.String())
				if notifyErr != nil {
					slog.Error("unable to send notification", "error", notifyErr)
				}
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
				notifyErr := notify.Publish(client, []byte(fmt.Sprintf("Failed to marshal the authentication token value recieved from the Whoop API. Additional context below: \n %s", err)), internal.EventErrors.String())
				if notifyErr != nil {
					slog.Error("unable to send notification", "error", notifyErr)
				}
				os.Exit(1)
			}

			err = os.WriteFile(config.Credentials.CredentialsFile, data, 0755)
			if err != nil {
				slog.Error("unable to write token file", "error", err)
				notifyErr := notify.Publish(client, []byte(fmt.Sprintf("Failed to write the authentication token value to the file. Additional context below: \n %s", err)), internal.EventErrors.String())
				if notifyErr != nil {
					slog.Error("unable to send notification", "error", notifyErr)
				}
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

			var ua string = UserAgent

			token, err := internal.ReadTokenFromFile(config.Credentials.CredentialsFile)
			if err != nil {
				slog.Error("unable to read token file", "error", err)
				notifyErr := notify.Publish(client, []byte(fmt.Sprintf("Failed to read the authentication token from file during the regular daily retreive cycle. Additional context below: \n %s", err)), internal.EventErrors.String())
				if notifyErr != nil {
					slog.Error("unable to send notification", "error", notifyErr)
				}
				os.Exit(1)
			}

			var user internal.User

			finalDataRaw, err := getData(ctx, user, client, token, &config.Server.FirstRunDownload, ua)
			if err != nil {
				slog.Error("unable to get data", "error", err)
				notifyErr := notify.Publish(client, []byte(fmt.Sprintf("Failed to get data from the Whoop API. Additional context below: \n %s", err)), internal.EventErrors.String())
				if notifyErr != nil {
					slog.Error("unable to send notification", "error", notifyErr)
				}
				os.Exit(1)
			}

			err = exp.Export(finalDataRaw)
			if err != nil {
				slog.Error("unable to export data", "error", err)
				notifyErr := notify.Publish(client, []byte(fmt.Sprintf("Failed to export data. Additional context below: \n %s", err)), internal.EventErrors.String())
				if notifyErr != nil {
					slog.Error("unable to send notification", "error", notifyErr)
				}
				os.Exit(1)
			}

			err = exp.CleanUp()
			if err != nil {
				slog.Error("unable to clean up export", "error", err)
				notifyErr := notify.Publish(client, []byte(fmt.Sprintf("Failed to clean up export. Additional context below: \n %s", err)), internal.EventErrors.String())
				if notifyErr != nil {
					slog.Error("unable to send notification", "error", notifyErr)
				}
				os.Exit(1)
			}

			slog.Info("Data collection complete")
			err = notify.Publish(client, []byte("Daily data collection complete."), internal.EventSuccess.String())
			if err != nil {
				slog.Error("unable to send notification", "error", err)
			}

		}

	}()

	return nil
}

// getData queries the Whoop API and gets the user data
func getData(ctx context.Context, user internal.User, client *http.Client, token oauth2.Token, firstDownload *bool, ua string) ([]byte, error) {

	if firstDownload == nil {
		slog.Debug("firstDownload is nil. Unable to determine if this is the first download")
		firstDownload = new(bool)
		*firstDownload = false
	}

	if !*firstDownload {
		startTime, endTime := internal.GenerateLast24HoursString()
		filterString := fmt.Sprintf("start=%s&end=%s", startTime, endTime)

		slog.Debug("Filter string", "filter", filterString)

		sleep, err := user.GetSleepCollection(ctx, client, internal.DEFAULT_WHOOP_API_USER_SLEEP_DATA_URL, token.AccessToken, filterString, ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		sleep.NextToken = ""
		user.SleepCollection = *sleep

		recovery, err := user.GetRecoveryCollection(ctx, client, internal.DEFAULT_WHOOP_API_RECOVERY_DATA_URL, token.AccessToken, filterString, ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		recovery.NextToken = ""
		user.RecoveryCollection = *recovery

		workout, err := user.GetWorkoutCollection(ctx, client, internal.DEFAULT_WHOOP_API_WORKOUT_DATA_URL, token.AccessToken, filterString, ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		workout.NextToken = ""
		user.WorkoutCollection = *workout

		cycle, err := user.GetCycleCollection(ctx, client, internal.DEFAULT_WHOOP_API_CYCLE_DATA_URL, token.AccessToken, filterString, ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		cycle.NextToken = ""
		user.CycleCollection = *cycle
	}

	if *firstDownload {

		data, err := user.GetUserProfileData(ctx, client, internal.DEFAULT_WHOOP_API_USER_DATA_URL, token.AccessToken, ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		user.UserData = *data

		measurements, err := user.GetUserMeasurements(ctx, client, internal.DEFAULT_WHOOP_API_USER_MEASUREMENT_DATA_URL, token.AccessToken, ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		user.UserMesaurements = *measurements

		sleep, err := user.GetSleepCollection(ctx, client, internal.DEFAULT_WHOOP_API_USER_SLEEP_DATA_URL, token.AccessToken, "", ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		sleep.NextToken = ""
		user.SleepCollection = *sleep

		recovery, err := user.GetRecoveryCollection(ctx, client, internal.DEFAULT_WHOOP_API_RECOVERY_DATA_URL, token.AccessToken, "", ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		recovery.NextToken = ""
		user.RecoveryCollection = *recovery

		workout, err := user.GetWorkoutCollection(ctx, client, internal.DEFAULT_WHOOP_API_WORKOUT_DATA_URL, token.AccessToken, "", ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		user.WorkoutCollection = *workout

		cycle, err := user.GetCycleCollection(ctx, client, internal.DEFAULT_WHOOP_API_CYCLE_DATA_URL, token.AccessToken, "", ua)
		if err != nil {
			internal.LogError(err)
			return []byte{}, err
		}

		cycle.NextToken = ""
		user.CycleCollection = *cycle

		// Set to false so that the entire data is not downloaded again
		*firstDownload = false
	}

	finalDataRaw, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		internal.LogError(err)
		return finalDataRaw, err
	}

	return finalDataRaw, nil

}
