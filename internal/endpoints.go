// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// getUserProfile returns the user profile from the Whoop API
func (u User) GetUserProfileData(ctx context.Context, client *http.Client, url, authToken, ua string) (*UserData, error) {

	const method = "GET"

	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		LogError(err)
		return nil, err
	}
	authHeader := "Bearer " + authToken
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("User-Agent", ua)

	response, err := client.Do(req)
	if err != nil {
		LogError(err)
		return nil, err
	}

	if response == nil {
		return nil, errors.New("the HTTP request for user profile data returned an empty response struct")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the HTTP request for user profile data returned an empty response struct or a non-200 status code. Status Code: %d", response.StatusCode)
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
		return nil, fmt.Errorf("unable to unmarshal data from Whoop API user profile payload."+"Received the follow errors: %w", err)
	}

	return &user, nil

}

// GetUserMeasurements returns the user measurements provided by the user from the Whoop API
func (u User) GetUserMeasurements(ctx context.Context, client *http.Client, url, authToken, ua string) (*UserMesaurements, error) {
	const method = "GET"

	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		LogError(err)
		return nil, err
	}
	authHeader := "Bearer " + authToken
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("User-Agent", ua)

	response, err := client.Do(req)
	if err != nil {
		LogError(err)
		return nil, err
	}
	if response == nil {
		return nil, errors.New("the HTTP request for user profile data returned an empty response struct")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the HTTP request for user profile data returned an empty response struct or a non-200 status code. Status Code: %d", response.StatusCode)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("unable to read response body from user mesurements payload", "msg", err)
		return nil, err
	}

	var user UserMesaurements
	err = json.Unmarshal(body, &user)
	if err != nil {
		slog.Error("unable to unmarshal data from Whoop API user mesaurement payload", "msg", err)
		return nil, err
	}

	return &user, nil

}

// GetSleepCollection returns the sleep collection from the Whoop API
// filters is a string of filters to apply to the request
// Pagination is enabled by default so as a result all available sleep collection records will be returned
func (u User) GetSleepCollection(ctx context.Context, client *http.Client, url, authToken, filters, ua string) (*SleepCollection, error) {
	const method = "GET"
	var sleep SleepCollection
	var sleepRecords []SleepCollectionRecords
	var continueLoop = true
	var nextLoopUrl string

	urlWithFilters := url

	if filters != "" {
		slog.Debug("Sleep Filters", slog.String("Filters", filters))
		urlWithFilters = url + filters
	}

	bo := generateBackoff()
	slog.Info(("Requesting sleep collection from Whoop API"))
	for continueLoop {

		if nextLoopUrl == "" {
			nextLoopUrl = urlWithFilters
		}
		slog.Debug("URL", slog.String("URL", nextLoopUrl))
		req, err := http.NewRequestWithContext(context.Background(), method, nextLoopUrl, nil)
		if err != nil {
			LogError(err)
			return nil, err
		}
		authHeader := "Bearer " + authToken
		req.Header.Add("Authorization", authHeader)
		req.Header.Add("User-Agent", ua)
		// Reset nextLoopUrl
		nextLoopUrl = ""

		op := func() error {

			response, err := client.Do(req)
			if err != nil {
				err = backoff.Permanent(err)
				LogError(err)
				return err
			}
			if response == nil {
				return errors.New("the HTTP request for user profile data returned an empty response struct")
			}

			defer response.Body.Close()

			if response.StatusCode == http.StatusTooManyRequests {
				slog.Info("Too many requests. Retrying...")
				return errors.New("too many requests")
			}

			if response.StatusCode > 400 && response.StatusCode <= 404 || response.StatusCode >= 500 {
				continueLoop = false
				err = fmt.Errorf("request errors related to authentication or server error. Status code is: %d", response.StatusCode)
				err = backoff.Permanent(err)
				return err
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				slog.Error("unable to read response body from sleep collection payload", "msg", err)
				return err
			}

			if len(body) == 0 {
				return errors.New("the Whoop API returned an empty response body for some sleep records. Retrying... ")
			}

			var sleep SleepCollection
			err = json.Unmarshal(body, &sleep)
			if err != nil {
				err = backoff.Permanent(err)
				LogError(err)
				return err
			}
			sleepRecords = append(sleepRecords, sleep.SleepCollectionRecords...)
			nextToken := sleep.NextToken

			if nextToken == nil {
				continueLoop = false
			} else {
				nextLoopUrl = urlWithFilters + "&nextToken=" + *nextToken
			}

			return nil

		}

		err = backoff.RetryNotify(op, bo, notification)
		if err != nil {
			return nil, err
		}
	}

	slog.Debug("Sleep Records", slog.Any("Sleep Records Count", len(sleepRecords)))

	sleep.SleepCollectionRecords = sleepRecords

	return &sleep, nil

}

func (u User) GetRecoveryCollection(ctx context.Context, client *http.Client, url, authToken, filters, ua string) (*RecoveryCollection, error) {

	const method = "GET"

	var recovery RecoveryCollection
	var recoveryRecords []RecoveryRecords
	var continueLoop = true
	var nextLoopUrl string

	urlWithFilters := url

	if filters != "" {
		urlWithFilters = url + filters
	}

	bo := generateBackoff()
	slog.Info(("Requesting recovery collection from Whoop API"))
	for continueLoop {

		if nextLoopUrl == "" {
			nextLoopUrl = urlWithFilters
		}
		slog.Debug("URL", slog.String("URL", nextLoopUrl))
		req, err := http.NewRequestWithContext(context.Background(), method, nextLoopUrl, nil)
		if err != nil {
			LogError(err)
			return nil, err
		}
		authHeader := "Bearer " + authToken
		req.Header.Add("Authorization", authHeader)
		req.Header.Add("User-Agent", ua)
		// Reset nextLoopUrl
		nextLoopUrl = ""

		op := func() error {

			response, err := client.Do(req)
			if err != nil {
				LogError(err)
				err = backoff.Permanent(err)
				return err
			}
			if response == nil {
				return errors.New("the HTTP request for recovery data returned an empty response struct")
			}

			defer response.Body.Close()

			if response.StatusCode == http.StatusTooManyRequests {
				slog.Info("Too many requests. Retrying...")
				return errors.New("too many requests")
			}

			if response.StatusCode > 400 && response.StatusCode <= 404 || response.StatusCode >= 500 {
				continueLoop = false
				err = fmt.Errorf("request errors related to authentication or server error. Status code is: %d", response.StatusCode)
				err = backoff.Permanent(err)
				return err
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				slog.Error("unable to read response body from recovery collection payload", "msg", err)
				return err
			}

			if len(body) == 0 {
				return errors.New("the Whoop API returned an empty response body for some recovery records. Retrying... ")
			}

			var recovery RecoveryCollection
			err = json.Unmarshal(body, &recovery)
			if err != nil {
				LogError(err)
				err = backoff.Permanent(err)
				return err
			}

			recoveryRecords = append(recoveryRecords, recovery.RecoveryRecords...)
			nextToken := recovery.NextToken

			if nextToken == nil {
				continueLoop = false
			} else {
				nextLoopUrl = urlWithFilters + "&nextToken=" + *nextToken
			}

			return nil

		}
		err = backoff.RetryNotify(op, bo, notification)
		if err != nil {
			return nil, err
		}
	}

	slog.Debug("Recovery Records", slog.Any("Recovery Records Count", len(recoveryRecords)))

	recovery.RecoveryRecords = recoveryRecords

	return &recovery, nil

}

func (u User) GetWorkoutCollection(ctx context.Context, client *http.Client, url, authToken, filters, ua string) (*WorkoutCollection, error) {

	const method = "GET"

	var workout WorkoutCollection
	var workoutRecords []WorkoutRecords
	var continueLoop = true
	var nextLoopUrl string

	urlWithFilters := url

	if filters != "" {
		urlWithFilters = url + filters
	}

	bo := generateBackoff()
	slog.Info(("Requesting workout collection from Whoop API"))
	for continueLoop {

		if nextLoopUrl == "" {
			nextLoopUrl = urlWithFilters
		}
		slog.Debug("URL", slog.String("URL", nextLoopUrl))
		req, err := http.NewRequestWithContext(context.Background(), method, nextLoopUrl, nil)
		if err != nil {
			LogError(err)
			return nil, err
		}
		authHeader := "Bearer " + authToken
		req.Header.Add("Authorization", authHeader)
		req.Header.Add("User-Agent", ua)
		// Reset nextLoopUrl
		nextLoopUrl = ""

		op := func() error {
			response, err := client.Do(req)
			if err != nil {
				LogError(err)
				err = backoff.Permanent(err)
				return err
			}

			if response == nil {
				return errors.New("the HTTP request for user profile data returned an empty response struct")
			}

			defer response.Body.Close()

			if response.StatusCode == http.StatusTooManyRequests {
				slog.Info("Too many requests. Retrying...")
				return errors.New("too many requests")
			}

			if response.StatusCode > 400 && response.StatusCode <= 404 || response.StatusCode >= 500 {
				continueLoop = false
				err = fmt.Errorf("request errors related to authentication or server error. Status code is: %d", response.StatusCode)
				err = backoff.Permanent(err)
				return err
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				slog.Error("unable to read response body from workout collection payload", "msg", err)
				err = backoff.Permanent(err)
				return err
			}

			if len(body) == 0 {
				return errors.New("the Whoop API returned an empty response body for some workout records. Retrying... ")
			}

			var workout WorkoutCollection
			err = json.Unmarshal(body, &workout)
			if err != nil {
				slog.Debug("Workout", "data", workout)
				LogError(err)
				err = backoff.Permanent(err)
				return err
			}

			workoutRecords = append(workoutRecords, workout.Records...)
			nextToken := workout.NextToken

			if nextToken == nil {
				continueLoop = false
			} else {
				nextLoopUrl = urlWithFilters + "&nextToken=" + *nextToken
			}

			return nil
		}

		err = backoff.RetryNotify(op, bo, notification)
		if err != nil {
			return nil, err
		}
	}

	slog.Debug("Workout Records", slog.Any("Workout Records Count", len(workoutRecords)))

	workout.Records = workoutRecords
	return &workout, nil
}

func (u User) GetCycleCollection(ctx context.Context, client *http.Client, url, authToken, filters, ua string) (*CycleCollection, error) {
	const method = "GET"

	var cycle CycleCollection
	var cycleRecords []CycleRecords
	var continueLoop = true
	var nextLoopUrl string

	urlWithFilters := url

	if filters != "" {
		urlWithFilters = url + filters
	}

	bo := generateBackoff()
	slog.Info(("Requesting cycle collection from Whoop API"))
	for continueLoop {

		if nextLoopUrl == "" {
			nextLoopUrl = urlWithFilters
		}
		slog.Debug("URL", slog.String("URL", nextLoopUrl))
		req, err := http.NewRequestWithContext(context.Background(), method, nextLoopUrl, nil)
		if err != nil {
			LogError(err)
			return nil, err
		}
		authHeader := "Bearer " + authToken
		req.Header.Add("Authorization", authHeader)
		req.Header.Add("User-Agent", ua)
		// Reset nextLoopUrl
		nextLoopUrl = ""

		op := func() error {
			response, err := client.Do(req)
			if err != nil {
				LogError(err)
				err = backoff.Permanent(err)
				return err
			}

			if response == nil {
				return errors.New("the HTTP request for cycle data returned an empty response struct")
			}

			defer response.Body.Close()

			if response.StatusCode > 400 && response.StatusCode <= 404 || response.StatusCode >= 500 {
				continueLoop = false
				err = backoff.Permanent(err)
				return err
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				slog.Error("unable to read response body from cycle collection payload", "msg", err)
				err = backoff.Permanent(err)
				return err
			}

			var currentCycle CycleCollection

			if len(body) == 0 {
				return errors.New("the Whoop API returned an empty response body for some cycle records. Retrying... ")
			}

			err = json.Unmarshal(body, &currentCycle)
			if err != nil {
				err = backoff.Permanent(err)
				LogError(err)
				return err
			}

			cycleRecords = append(cycleRecords, currentCycle.Records...)
			nextToken := currentCycle.NextToken

			if nextToken == nil {
				continueLoop = false
			} else {
				nextLoopUrl = urlWithFilters + "&nextToken=" + *nextToken
			}

			return nil
		}

		err = backoff.RetryNotify(op, bo, notification)
		if err != nil {
			return nil, err
		}
	}

	slog.Debug("Cycle Records", slog.Any("Cycle Records Count", len(cycleRecords)))

	cycle.Records = cycleRecords
	return &cycle, nil
}

// notification is the default notification logic for the backoff retryer
func notification(err error, duration time.Duration) {

	if err != nil {
		slog.Error("error", "msg", err)
	}

	slog.Info("Error getting sleep records", "Retrying in: ", duration.String())

}

// generateBackoff returns a backoff exponential backoff struct
func generateBackoff() *backoff.ExponentialBackOff {

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = DEFAULT_RETRY_INITIAL_INTERVAL
	bo.Multiplier = DEFAULT_RETRY_MULTIPLIER
	bo.RandomizationFactor = DEFAULT_RETRY_RANDOMIZATION
	bo.MaxElapsedTime = DEFAULT_RETRY_MAX_ELAPSED_TIME

	return bo
}
