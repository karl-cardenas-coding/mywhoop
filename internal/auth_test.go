// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestGetEndpoint(t *testing.T) {

	expected := oauth2.Endpoint{
		AuthURL:  DEFAULT_AUTHENTICATION_URL,
		TokenURL: DEFAULT_ACCESS_TOKEN_URL,
	}

	got := getEndpoint()

	if got.AuthURL != expected.AuthURL {
		t.Errorf("an error occured. Expected %s but got %s", expected.AuthURL, got.AuthURL)
	}

	if got.TokenURL != expected.TokenURL {
		t.Errorf("an error occured. Expected %s but got %s", expected.TokenURL, got.TokenURL)
	}

}

func TestWriteLocalToken(t *testing.T) {

	tests := []struct {
		id                    int
		token                 oauth2.Token
		tokenPath             string
		errorExpected         bool
		createDirectoryBefore bool
		createFileBefore      bool
		checkFileCreated      bool
	}{
		{
			0,
			oauth2.Token{
				AccessToken:  "askjdsajklsdlkjfasdk",
				TokenType:    "Bearer",
				RefreshToken: "4ssdjfdsjokdfsopfdopkdfsjksdjiujiosaoi",
				// Expires in 30 min
				Expiry: time.Now().Add(30 * time.Minute),
			},
			"../tests/data/",
			false,
			true,
			false,
			true,
		},
		{
			0,
			oauth2.Token{
				AccessToken:  "askjdsajklsdlkjfasdk",
				TokenType:    "Bearer",
				RefreshToken: "4ssdjfdsjokdfsopfdopkdfsjksdjiujiosaoi",
				// Expires in 30 min
				Expiry: time.Now().Add(30 * time.Minute),
			},
			"../tests/data/",
			true,
			true,
			true,
			false,
		},
	}

	for index, test := range tests {
		test.id = index + 1

		if test.createDirectoryBefore {
			err := os.MkdirAll(test.tokenPath, 0755)
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create directory: %v", test.id, err)
			}
		}

		if test.createFileBefore {

			file, err := os.Create(filepath.Join(test.tokenPath, "token.json"))
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create file: %v", test.id, err)
			}

			err = file.Close()
			if err != nil {
				t.Errorf("Test Case - %d: Failed to close file: %v", test.id, err)
			}

		}

		filePath := filepath.Join(test.tokenPath, "token.json")

		err := writeLocalToken(filePath, &test.token)
		if !test.errorExpected && err != nil {
			t.Errorf("Test Case - %d: Failed to write token to file: %v", test.id, err)
		}

		if test.checkFileCreated {
			_, err := os.Stat(filePath)
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create file: %v", test.id, err)
			}
		}

		err = cleanUp("../tests/data/")
		if err != nil {
			t.Errorf("Test Case - %d: Failed to clean up: %v", test.id, err)
		}

	}

}

func TestReadTokenFromFile(t *testing.T) {

	type mockToken struct {
		AccessToken  string    `json:"access_token"`
		TokenType    string    `json:"token_type"`
		RefreshToken string    `json:"refresh_token"`
		Scopes       string    `json:"scopes"`
		Expiry       time.Time `json:"expiry"`
	}

	invalid := "invalid"

	tests := []struct {
		id                    int
		token                 mockToken
		tokenPath             string
		errorExpected         bool
		createDirectoryBefore bool
		createFileBefore      bool
		checkFileCreated      bool
		customContent         bool
	}{
		// Test case for valid token
		{
			0,
			mockToken{
				AccessToken:  "lQZyVtOv_d3QV6_-baV2Ffskbx3jqlsNfioTkXnVQpM.m9W5lWsF8ALeumQdnHibh8-bc3IYkdXGu8qsz23VSng",
				TokenType:    "bearer",
				RefreshToken: "4ssdjfdsjokdfsopfdopkdfsjksdjiujiosaoi",
				Scopes:       "offline read:profile read:recovery read:cycles read:workout read:sleep read:body_measurement",
				Expiry:       time.Now().Local().Add(time.Second * time.Duration(1800)),
			},
			"../tests/data/",
			false,
			true,
			true,
			true,
			false,
		},
		// Test case for invalid token
		{
			0,
			mockToken{},
			"../tests/data/",
			true,
			true,
			true,
			true,
			true,
		},
		// Test case for no token file available
		{
			0,
			mockToken{},
			"../tests/data/",
			true,
			false,
			false,
			false,
			false,
		},
	}

	for index, test := range tests {
		test.id = index + 1

		if test.createDirectoryBefore {
			err := os.MkdirAll(test.tokenPath, 0755)
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create directory: %v", test.id, err)
			}
		}

		filePath := filepath.Join(test.tokenPath, "token.json")

		if test.createFileBefore {

			f, err := os.Create(filePath)
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create file: %v", test.id, err)
			}

			var rs []byte
			if test.customContent {
				value, err := json.MarshalIndent(invalid, " ", " ")
				if err != nil {
					t.Errorf("Test Case - %d: Failed to marshal token: %v", test.id, err)
				}
				rs = value

			} else {
				value, err := json.MarshalIndent(test.token, " ", " ")
				if err != nil {
					t.Errorf("Test Case - %d: Failed to marshal token: %v", test.id, err)
				}
				rs = value

			}

			_, err = f.WriteString(string(rs))
			if err != nil {
				t.Errorf("Test Case - %d: Failed to write to file: %v", test.id, err)
			}

			err = f.Close()
			if err != nil {
				t.Errorf("Test Case - %d: Failed to close file: %v", test.id, err)
			}

		}

		token, err := ReadTokenFromFile(filePath)

		if err != nil && !test.errorExpected {
			t.Errorf("Test Case - %d: Failed to read token from file: %v", test.id, err)
		}

		if err == nil && test.errorExpected {
			t.Errorf("Test Case - %d: Expected an error but got none. Produced this token: %v", test.id, token)

		}

		if token.AccessToken != test.token.AccessToken {
			t.Errorf("Test Case - %d: Expected %s but got %s", test.id, test.token.AccessToken, token.AccessToken)
		}

		if token.RefreshToken != test.token.RefreshToken {
			t.Errorf("Test Case - %d: Expected %s but got %s", test.id, test.token.RefreshToken, token.RefreshToken)
		}

		if token.TokenType != test.token.TokenType {
			t.Errorf("Test Case - %d: Expected %s but got %s", test.id, test.token.TokenType, token.TokenType)
		}

		if token.Expiry != test.token.Expiry {
			t.Errorf("Test Case - %d: Expected %s but got %s", test.id, test.token.Expiry, token.Expiry)
		}

		err = cleanUp("../tests/data/")
		if err != nil {
			t.Errorf("Test Case - %d: Failed to clean up: %v", test.id, err)
		}

	}

}

// cleanUp removes the tests directory
func cleanUp(path string) error {

	var pathToDelete string

	currentDir, err := os.Getwd()
	if err != nil {

		return err
	}

	pathToDelete = filepath.Join(currentDir, "tests", "data")

	if path != "" {
		pathToDelete = path
	}

	err = os.RemoveAll(pathToDelete)
	if err != nil {
		slog.Error("Failed to remove directory", "msg", err)
		err := exec.Command("rm -rf export/tests/").Run()
		if err != nil {
			return err
		}
		return err
	}

	return nil
}
