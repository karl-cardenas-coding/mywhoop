// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"os"
)

// checkRequiredEnvVars checks if the required environment variables are set
// A ConfigurationData struct is returned with the values of the environment variables
func ExtractEnvVariables() (ConfigurationData, error) {

	var output ConfigurationData

	id := os.Getenv("WHOOP_CLIENT_ID")
	secret := os.Getenv("WHOOP_CLIENT_SECRET")

	credsFile := os.Getenv("WHOOP_CREDENTIALS_FILE")

	switch {
	case id == "" && secret == "":
		return output, errors.New("the required env variables WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET are not set")

	case id == "":
		return output, errors.New("the required env variable WHOOP_CLIENT_ID is not set")

	case secret == "":
		return output, errors.New("the required env variable WHOOP_CLIENT_SECRET is not set")

	}

	if credsFile != "" {
		output.Credentials.CredentialsFile = credsFile
	}

	return output, nil
}
