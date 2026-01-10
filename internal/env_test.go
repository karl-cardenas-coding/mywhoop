// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"testing"
)

func TestExtractEnvVariablesError(t *testing.T) {

	_ = os.Unsetenv("WHOOP_CLIENT_ID")
	_ = os.Unsetenv("WHOOP_CLIENT_SECRET")
	_ = os.Unsetenv("WHOOP_CREDENTIALS_FILE")

	expectedMsg := "the required env variables WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET are not set"
	_, err := ExtractEnvVariables()
	if err == nil {
		t.Log(err)
		t.Errorf("Expected  an error, but got %v", err)
	}

	if err.Error() != expectedMsg {
		t.Errorf("Expected the following error message:  %s - but got: %s", expectedMsg, err.Error())
	}

	cleanUpEnvVars()
}

func TestExtractEnvVariablesClient(t *testing.T) {

	cleanUpEnvVars()

	_ = os.Setenv("WHOOP_CLIENT_ID", "AAAAAAAAAAAAAAAAAAA")
	expectedMsg := "the required env variable WHOOP_CLIENT_SECRET is not set"
	_, err := ExtractEnvVariables()

	if err == nil {
		t.Errorf("Expected  an error, but got %v", err)
	}

	if err.Error() != expectedMsg {
		t.Errorf("Expected the following error message:  %s - but got: %s", expectedMsg, err.Error())
	}

	cleanUpEnvVars()
}

func TestExtractEnvVariablesClientSecret(t *testing.T) {

	cleanUpEnvVars()

	_ = os.Setenv("WHOOP_CLIENT_SECRET", "BBBBBBBBBBBBBBBBBBBBB")
	expectedMsg := "the required env variable WHOOP_CLIENT_ID is not set"
	_, err := ExtractEnvVariables()

	if err == nil {
		t.Log(err)
		t.Errorf("Expected  an error, but got %v", err)
	}

	if err.Error() != expectedMsg {
		t.Errorf("Expected the following error message:  %s - but got: %s", expectedMsg, err.Error())
	}

	cleanUpEnvVars()

}

func TestExtractEnvVariablesCredsFileEmpty(t *testing.T) {

	_ = os.Setenv("WHOOP_CLIENT_ID", "AAAAAAAAAAAAAAAAAAA")
	_ = os.Setenv("WHOOP_CLIENT_SECRET", "BBBBBBBBBBBBBBBBBBBBB")

	cfg, err := ExtractEnvVariables()
	if err != nil {
		t.Log(err)
		t.Errorf("Expected no error but got %v", err)
	}

	if cfg.Credentials.CredentialsFile != "" {
		t.Errorf("Expected an empty variable but got %v", cfg.Credentials.CredentialsFile)
	}

	cleanUpEnvVars()

}

func TestExtractEnvVariablesCredsFile(t *testing.T) {

	_ = os.Setenv("WHOOP_CLIENT_ID", "AAAAAAAAAAAAAAAAAAA")
	_ = os.Setenv("WHOOP_CLIENT_SECRET", "BBBBBBBBBBBBBBBBBBBBB")
	_ = os.Setenv("WHOOP_CREDENTIALS_FILE", "myToken.json")

	cfg, err := ExtractEnvVariables()
	if err != nil {
		t.Log(err)
		t.Errorf("Expected no error but got %v", err)
	}

	if cfg.Credentials.CredentialsFile != "myToken.json" {
		t.Errorf("Expected an empty variable but got %v", cfg.Credentials.CredentialsFile)
	}

	cleanUpEnvVars()

}

// cleanUpEnvVars unsets the required env variables.
func cleanUpEnvVars() {
	_ = os.Unsetenv("WHOOP_CLIENT_ID")
	_ = os.Unsetenv("WHOOP_CLIENT_SECRET")
	_ = os.Unsetenv("WHOOP_CREDENTIALS_FILE")
}
