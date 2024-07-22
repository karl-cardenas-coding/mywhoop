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
	"strings"
	"syscall"
	"time"

	gocron "github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/karl-cardenas-coding/mywhoop/notifications"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
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

func init() {
	rootCmd.AddCommand(serverCmd)
}

// EvaluateConfigOptions evaluates the configuration options for the server command
// Command line options take precedence over configuration file options.
func evaluateConfigOptions(cfg *internal.ConfigurationData) error {

	if cfg.Export.Method == "" {
		slog.Info("No exporter specified. Defaulting to file.")
		cfg.Export.Method = "file"
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

	// Evaluate the configuration options
	err = evaluateConfigOptions(&cfg)
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
		return errors.New("unknown exporter")
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

	sch, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local),
		gocron.WithLogger(
			gocron.NewLogger(loggerConverter(cfg.Debug)),
		),
	)
	if err != nil {
		slog.Error("unable to create scheduler", "error", err)
		return err
	}
	defer func() {
		err = sch.Shutdown()
		if err != nil {
			slog.Error("unable to shutdown scheduler", "error", err)
		}
		os.Exit(1)
	}()

	// This job is to refresh the token immediately upon startup
	// This is to ensure that the token is valid. If the token is invalid, the server will exit and the user will be notified immediately upon startup.
	_, err = sch.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartImmediately(),
		),
		gocron.NewTask(func() error {
			slog.Info("Refreshing auth token token")
			return refreshJWT(ctx, client, cfg.Credentials.CredentialsFile)
		}),
		gocron.WithName("mywhoop_startup_token_refresh_job"),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithEventListeners(
			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, err error) {
					slog.Error("error completing the startup token refresh job", "error", err)
					notifyErr := notificationMethod.Publish(client, []byte(fmt.Sprintf("Error running the token refresh job. Additional context below: \n %s", err)), internal.EventErrors.String())
					if notifyErr != nil {
						slog.Error("unable to send notification", "error", notifyErr)
					}
					os.Exit(1)
				},
			),
		),
	)
	if err != nil {
		slog.Error("unable to create the immediate one-time JWT refresh upon startup", "error", err)
		return err
	}

	_, err = sch.NewJob(
		gocron.DurationJob(
			jwtRefreshDurationValidator(cfg.Server.JWTRefreshDuration),
		),
		gocron.NewTask(func() error {
			slog.Info("Refreshing auth token token")
			return refreshJWT(ctx, client, cfg.Credentials.CredentialsFile)
		}),
		gocron.WithName("mywhoop_token_refresh_job"),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithEventListeners(
			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, err error) {
					slog.Error("error completing the token refresh job", "error", err)
					notifyErr := notificationMethod.Publish(client, []byte(fmt.Sprintf("Error running the token refresh job. Additional context below: \n %s", err)), internal.EventErrors.String())
					if notifyErr != nil {
						slog.Error("unable to send notification", "error", notifyErr)
					}
					os.Exit(1)
				},
			),
		),
	)
	if err != nil {
		slog.Error("unable to create token cron job", "error", err)
		return err
	}

	var cronValue string
	if cfg.Server.Crontab != "" {
		cronValue = cfg.Server.Crontab
	} else {
		cronValue = internal.DEFAULT_SERVER_CRON_SCHEDULE
	}
	slog.Debug("Cron schedule", "schedule", cronValue)

	_, err = sch.NewJob(
		gocron.CronJob(cronValue, false),
		gocron.NewTask(downloadWhoopData,
			ctx,
			cfg,
			client,
			exportSelected,
			notificationMethod),
		gocron.WithName("mywhoop_data_collection_job"),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithEventListeners(
			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, err error) {
					slog.Error("error running server job", "error", err)
					notifyErr := notificationMethod.Publish(client, []byte(fmt.Sprintf("Error running the server job. Additional context below: \n %s", err)), internal.EventErrors.String())
					if notifyErr != nil {
						slog.Error("unable to send notification", "error", notifyErr)
					}
					os.Exit(1)
				},
			),
		),
	)
	if err != nil {
		slog.Error("unable to create cron job", "error", err)
		return err
	}

	sch.Start()

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
		err = sch.StopJobs()
		if err != nil {
			slog.Error("unable to stop jobs", "error", err)
			notifyErr := notificationMethod.Publish(client, []byte(fmt.Sprintf("unable to stop jobs. Additional error message: \n %s", err)), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
		}
		err = sch.Shutdown()
		if err != nil {
			slog.Error("unable to shutdown scheduler", "error", err)
			notifyErr := notificationMethod.Publish(client, []byte(fmt.Sprintf("unable to shutdown scheduler. Additional error message: \n %s", err)), internal.EventErrors.String())
			if notifyErr != nil {
				slog.Error("unable to send notification", "error", notifyErr)
			}
		}
		slog.Info("Server shutdown complete")
		os.Exit(0)
	}

	return nil
}

// downloadWhoopData orchestrates the download of Whoop data and ensures the data is exported.
func downloadWhoopData(ctx context.Context, config internal.ConfigurationData, client *http.Client, exp internal.Export, notify internal.Notification) error {

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

	finalDataRaw, err := getData(ctx, user, client, token, ua)
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

	return nil
}

// refreshJWT refreshes the Whoop API JWT token.
func refreshJWT(ctx context.Context, client *http.Client, credentialsFilePath string) error {
	currentToken, err := internal.ReadTokenFromFile(credentialsFilePath)
	if err != nil {
		return err
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
		return err
	}

	if len(token.AccessToken) < 1 {
		return errors.New("no access token")
	}

	slog.Info("New token generated:", token.AccessToken[0:4], "....")

	data, err := json.MarshalIndent(token, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile(credentialsFilePath, data, 0755)
	if err != nil {
		return err
	}

	return nil
}

// getData queries the Whoop API and gets the user data
func getData(ctx context.Context, user internal.User, client *http.Client, token oauth2.Token, ua string) ([]byte, error) {

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

	finalDataRaw, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		internal.LogError(err)
		return finalDataRaw, err
	}

	return finalDataRaw, nil

}

// loggerConverter converts the string log level to the gocron log level
func loggerConverter(lvl string) gocron.LogLevel {
	switch strings.ToLower(lvl) {
	case "debug":
		return gocron.LogLevelDebug
	case "info":
		return gocron.LogLevelInfo
	case "warn":
		return gocron.LogLevelWarn
	case "error":
		return gocron.LogLevelError
	default:
		return gocron.LogLevelInfo
	}
}

// jwtRefreshDurationValidator validates the JWT refresh duration.
// If the duration is greater than 59 minutes or less than 0, the default DEFAULT_SERVER_TOKEN_REFRESH_CRON_SCHEDULE is used.
func jwtRefreshDurationValidator(incoming int) time.Duration {

	parsedValue := time.Duration(incoming) * time.Minute

	if parsedValue > 59*time.Minute || parsedValue <= 0*time.Minute {
		return internal.DEFAULT_SERVER_TOKEN_REFRESH_CRON_SCHEDULE
	}

	slog.Debug("JWT Refresh Duration", "duration", parsedValue)
	return parsedValue

}
