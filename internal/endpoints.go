package internal

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
)

// getUserProfile returns the user profile from the Whoop API
func (u User) GetData(ctx context.Context, client *http.Client, authToken string) (*UserData, error) {

	const (
		url    = "https://api.prod.whoop.com/developer/v1/user/profile/basic"
		method = "GET"
	)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		LogError(err)
		return nil, err
	}
	authHeader := "Bearer " + authToken
	req.Header.Add("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		LogError(err)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		LogError(err)
		return nil, err
	}

	var user UserData
	err = json.Unmarshal(body, &user)
	if err != nil {
		LogError(err)
		return nil, err
	}

	return &user, nil

}

func (u User) GetUserMeasurements(ctx context.Context, client *http.Client, authToken string) (*UserMesaurements, error) {
	const (
		url    = "https://api.prod.whoop.com/developer/v1/user/measurement/body"
		method = "GET"
	)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		LogError(err)
		return nil, err
	}
	authHeader := "Bearer " + authToken
	req.Header.Add("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		LogError(err)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("unable to read response body from user mesurements payload", err)
		return nil, err
	}

	var user UserMesaurements
	err = json.Unmarshal(body, &user)
	if err != nil {
		slog.Error("unable to unmarshal data from Whoop API user mesaurement payload", err)
		return nil, err
	}

	return &user, nil

}

// GetSleepCollection returns the sleep collection from the Whoop API
// filters is a string of filters to apply to the request
// Pagination is enabled by default so as a result all available sleep collection records will be returned
func (u User) GetSleepCollection(ctx context.Context, client *http.Client, authToken string, filters string) (*SleepCollection, error) {
	const (
		url    = "https://api.prod.whoop.com/developer/v1/activity/sleep?"
		method = "GET"
	)

	var sleep SleepCollection
	var sleepRecords []SleepCollectionRecords
	var continueLoop = true
	var nextLoopUrl string

	urlWithFilters := url

	if filters != "" {
		urlWithFilters = url + filters
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 500 * time.Millisecond
	bo.Multiplier = 1.5
	bo.RandomizationFactor = 0.5

	for continueLoop {

		slog.Info(("Requesting sleep collection from Whoop API"))
		slog.Debug("URL", slog.String("URL", urlWithFilters))

		if nextLoopUrl == "" {
			nextLoopUrl = urlWithFilters
		}
		req, err := http.NewRequestWithContext(ctx, method, nextLoopUrl, nil)
		if err != nil {
			LogError(err)
			return nil, err
		}
		authHeader := "Bearer " + authToken
		req.Header.Add("Authorization", authHeader)
		// Reset nextLoopUrl
		nextLoopUrl = ""

		err = backoff.RetryNotify(func() error {

			response, err := client.Do(req)
			if err != nil {
				LogError(err)
				return err
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				slog.Error("unable to read response body from sleep collection payload", err)
				return err
			}

			var sleep SleepCollection
			err = json.Unmarshal(body, &sleep)
			if err != nil {
				LogError(err)
				return err
			}
			sleepRecords = append(sleepRecords, sleep.SleepCollectionRecords...)
			nextToken := sleep.NextToken

			if nextToken == "" {
				continueLoop = false
			} else {
				nextLoopUrl = urlWithFilters + "nextToken=" + nextToken
			}

			return err

		}, bo, func(err error, duration time.Duration) {
			slog.Info("Too many requests. Error getting registries", "Retrying in: ", duration.String())
		})

		if err != nil {
			return nil, err
		}
	}

	sleep.SleepCollectionRecords = sleepRecords

	return &sleep, nil

}
