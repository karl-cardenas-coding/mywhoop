// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import "time"

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
)
