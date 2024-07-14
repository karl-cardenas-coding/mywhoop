// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"time"
)

const (
	// AccessTokenURL is the URL to get an access token from the Whoop API
	DEFAULT_ACCESS_TOKEN_URL string = "https://api.prod.whoop.com/oauth/oauth2/token"
	// AuthURL is the URL to authenticate with the Whoop API
	DEFAULT_AUTHENTICATION_URL string = "https://api.prod.whoop.com/oauth/oauth2/auth"
	// DEFAULT_REDIRECT_URL is the URL to redirect to after authentication
	DEFAULT_REDIRECT_URL string = "http://localhost:8080/oauth/redirect"
	// DEFAULT_CREDENTIALS_FILE is the default file to store the credentials
	DEFAULT_CREDENTIALS_FILE string = "token.json"
	// DEFAULT_CONFIG_FILE is the default file to store the configuration
	DEFAULT_CONFIG_FILE string = ".mywhoop.yaml"
	// Retry/Backoff constants
	DEFAULT_RETRY_MAX_ELAPSED_TIME time.Duration = 5 * time.Minute
	DEFAULT_RETRY_MULTIPLIER       float64       = 1.5
	DEFAULT_RETRY_RANDOMIZATION    float64       = 0.5
	DEFAULT_RETRY_INITIAL_INTERVAL time.Duration = 500 * time.Millisecond
	// DEFAULT_WHOOP_API_USER_DATA_URL is the URL to get the user data from the Whoop API
	DEFAULT_WHOOP_API_USER_DATA_URL = "https://api.prod.whoop.com/developer/v1/user/profile/basic"
	// DEFAULT_WHOOP_API_USER_MEASUREMENT_DATA_URL is the URL to get the user measurement data from the Whoop API
	DEFAULT_WHOOP_API_USER_MEASUREMENT_DATA_URL = "https://api.prod.whoop.com/developer/v1/user/measurement/body"
	// DEFAULT_WHOOP_API_USER_SLEEP_DATA_URL is the URL to get the user sleep data from the Whoop API
	DEFAULT_WHOOP_API_USER_SLEEP_DATA_URL = "https://api.prod.whoop.com/developer/v1/activity/sleep?"
	// DEFAULT_WHOPP_API_RECOVERY_DATA_URL is the URL to get the user recovery data from the Whoop API
	DEFAULT_WHOOP_API_RECOVERY_DATA_URL = "https://api.prod.whoop.com/developer/v1/recovery?"
	//DEFAULT_WHOPP_API_WORKOUT_DATA_URL is the URL to get the user workout data from the Whoop API
	DEFAULT_WHOOP_API_WORKOUT_DATA_URL = "https://api.prod.whoop.com/developer/v1/activity/workout?"
	// DEFAULT_WHOOP_API_CYCLE_DATA_URL is the URL to get the user cycle data from the Whoop API
	DEFAULT_WHOOP_API_CYCLE_DATA_URL = "https://api.prod.whoop.com/developer/v1/cycle?"
	// DEFAULT_SERVER_CRON_SCHEDULE is the default cron schedule for the server. Everyday at 1:00 PM OR 1300 hours.
	DEFAULT_SERVER_CRON_SCHEDULE string = "0 13 * * *"
	// DEFAULT_SERVER_TOKEN_REFRESH_CRON_SCHEDULE is the default cron schedule for the token refresh. Every 45 minutes.
	DEFAULT_SERVER_TOKEN_REFRESH_CRON_SCHEDULE time.Duration = 45 * time.Minute
)
