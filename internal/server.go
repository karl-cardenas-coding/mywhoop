package internal

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"golang.org/x/exp/slog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// StartServer starts the long running server.
func StartServer(ctx context.Context, config ConfigurationData, client *http.Client) error {

	ok, _, err := verfyToken("token.json")
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
			currentToken, err := readTokenFromFile("token.json")
			if err != nil {
				slog.Error("unable to read token file", "error", err)
				os.Exit(1)
			}

			token, err := RefreshToken(ctx, currentToken.AccessToken, currentToken.RefreshToken, client)
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
			slog.Info(auth.AccessToken)

			data, err := json.MarshalIndent(auth, "", " ")
			if err != nil {
				slog.Error("unable to marshal token", "error", err)
				os.Exit(1)
			}

			err = os.WriteFile("token.json", data, 0755)
			if err != nil {
				slog.Error("unable to write token file", "error", err)
				os.Exit(1)
			}
		}
	}()

	// This goroutine queries the Whoop API 24 hrs.
	go func() {

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {

			slog.Info("Starting data collection")

			token, err := readTokenFromFile("token.json")
			if err != nil {
				slog.Error("unable to read token file", "error", err)
				os.Exit(1)
			}

			var user User
			data, err := user.GetUserProfileData(ctx, client, token.AccessToken)
			if err != nil {
				LogError(err)
			}

			user.UserData = *data

			measurements, err := user.GetUserMeasurements(ctx, client, token.AccessToken)
			if err != nil {
				LogError(err)
			}

			user.UserMesaurements = *measurements

			sleep, err := user.GetSleepCollection(ctx, client, token.AccessToken, "")
			if err != nil {
				LogError(err)
			}

			user.SleepCollection = *sleep

			recovery, err := user.GetRecoveryCollection(ctx, client, token.AccessToken, "")
			if err != nil {
				LogError(err)
			}

			user.RecoveryCollection = *recovery

			workout, err := user.GetWorkoutCollection(ctx, client, token.AccessToken, "")
			if err != nil {
				LogError(err)
			}

			user.WorkoutCollection = *workout

			finalDataRaw, err := json.MarshalIndent(user, "", "  ")
			if err != nil {
				LogError(err)
			}

			fileExp := export.FileExport{
				FilePath: config.Export.FileExport.FilePath,
				FileType: config.Export.FileExport.FileType,
				FileName: config.Export.FileExport.FileName,
			}

			switch config.Export.Method {
			case "file":
				err = fileExp.Export(finalDataRaw)
				if err != nil {
					LogError(err)
				}
			default:
				err = fileExp.Export(finalDataRaw)
				if err != nil {
					LogError(err)
				}

			}

			slog.Info("Data collection complete")
		}

	}()

	return nil
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
