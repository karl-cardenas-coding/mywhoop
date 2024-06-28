// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestRefreshToken(t *testing.T) {

	client := CreateHTTPClient()

	type expected struct {
		AccessToken  string
		RefreshToken string
		TokenType    string
		Scope        string
		ExperiesIn   int
	}

	tests := []struct {
		id            int
		auth          AuthRequest
		exp           expected
		ts            *httptest.Server
		errorExpected bool
	}{
		{
			0,
			AuthRequest{
				AuthToken:    DEFAULT_ACCESS_TOKEN_URL,
				ClientID:     "testClientID",
				ClientSecret: "testClientSecret",
				RefreshToken: "testRefresh",
				TokenURL:     DEFAULT_ACCESS_TOKEN_URL,
				Client:       client,
			},
			expected{
				AccessToken:  "jkdjkasdsiaoiasoiuashnbnxgyyd4.G43yAmvMpGI8R5d_3MYVM8N0xFSCLOrAB2sGNUgl9U0",
				TokenType:    "bearer",
				Scope:        "offline read:profile read:recovery read:cycles read:workout read:sleep read:body_measurement",
				ExperiesIn:   3599,
				RefreshToken: "jfhkjfdkhjfdskljlkccxcxsdezzZ.uwypCEcZDN3zK-4PjSTrT9nADIE5AJGxYgs8FvYVx18",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"access_token": "jkdjkasdsiaoiasoiuashnbnxgyyd4.G43yAmvMpGI8R5d_3MYVM8N0xFSCLOrAB2sGNUgl9U0",
					"expires_in": 3599,
					"refresh_token": "jfhkjfdkhjfdskljlkccxcxsdezzZ.uwypCEcZDN3zK-4PjSTrT9nADIE5AJGxYgs8FvYVx18",
					"scope": "offline read:profile read:recovery read:cycles read:workout read:sleep read:body_measurement",
					"token_type": "bearer"
				}`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0,
			AuthRequest{
				AuthToken:    DEFAULT_ACCESS_TOKEN_URL,
				ClientID:     "testClientID",
				ClientSecret: "testClientSecret",
				RefreshToken: "testRefresh",
				TokenURL:     DEFAULT_ACCESS_TOKEN_URL,
				Client:       client,
			},
			expected{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`11212154544545`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			true,
		},
	}

	for index, test := range tests {
		test.id = index + 1
		defer test.ts.Close()
		ctx := context.Background()

		test.auth.TokenURL = test.ts.URL

		token, err := RefreshToken(ctx, test.auth)
		if err != nil && !test.errorExpected {
			t.Errorf("Test Case - %d: Failed to refresh token: %v", test.id, err)
		}

		fmt.Println(token)

		if err == nil && test.errorExpected {
			t.Errorf("Test Case - %d: Expected an error but got none", test.id)
		}

		if token.AccessToken != test.exp.AccessToken {
			t.Errorf("Test Case - %d: Expected %s but got %s", test.id, test.exp.AccessToken, token.AccessToken)
		}

		if token.RefreshToken != test.exp.RefreshToken {
			t.Errorf("Test Case - %d: Expected %s but got %s", test.id, test.exp.RefreshToken, token.RefreshToken)
		}

		if token.TokenType != test.exp.TokenType {
			t.Errorf("Test Case - %d: Expected %s but got %s", test.id, test.exp.TokenType, token.TokenType)
		}

		// An empty OAuth2 token return with an expiration value. This checks to make sure we are not dealing with an empty token
		if token.AccessToken != "" && token.RefreshToken != "" {

			expiresIn := int(time.Until(token.Expiry).Seconds()) + 1

			if expiresIn != test.exp.ExperiesIn {
				t.Errorf("Test Case - %d: Expires -  Expected %d but got %d", test.id, test.exp.ExperiesIn, expiresIn)
			}

		}

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

		err := WriteLocalToken(filePath, &test.token)
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
		description           string
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
			"Test Case - 1: Valid token",
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
			"Test Case - 2: Invalid token",
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
			"Test Case - 3: No token file available",
			mockToken{},
			"../tests/data/",
			true,
			false,
			false,
			false,
			false,
		},
	}

	for _, test := range tests {

		t.Run(test.description, func(t *testing.T) {

			if test.createDirectoryBefore {
				err := os.MkdirAll(test.tokenPath, 0755)
				if err != nil {
					t.Errorf("%s: Failed to create directory: %v", test.description, err)
				}
			}

			filePath := filepath.Join(test.tokenPath, "token.json")

			if test.createFileBefore {

				f, err := os.Create(filePath)
				if err != nil {
					t.Errorf("%s: Failed to create file: %v", test.description, err)
				}

				var rs []byte
				if test.customContent {
					value, err := json.MarshalIndent(invalid, " ", " ")
					if err != nil {
						t.Errorf("%s: Failed to marshal token: %v", test.description, err)
					}
					rs = value

				} else {
					value, err := json.MarshalIndent(test.token, " ", " ")
					if err != nil {
						t.Errorf("%s: Failed to marshal token: %v", test.description, err)
					}
					rs = value

				}

				_, err = f.WriteString(string(rs))
				if err != nil {
					t.Errorf("%s: Failed to write to file: %v", test.description, err)
				}

				err = f.Close()
				if err != nil {
					t.Errorf("%s: Failed to close file: %v", test.description, err)
				}

			}

			token, err := ReadTokenFromFile(filePath)

			if err != nil && !test.errorExpected {
				t.Errorf("%s: Failed to read token from file: %v", test.description, err)
			}

			if err == nil && test.errorExpected {
				t.Errorf("%s: Expected an error but got none. Produced this token: %v", test.description, token)

			}

			token.AccessToken = strings.TrimSpace(token.AccessToken)
			token.RefreshToken = strings.TrimSpace(token.RefreshToken)
			test.token.AccessToken = strings.TrimSpace(test.token.AccessToken)
			test.token.RefreshToken = strings.TrimSpace(test.token.RefreshToken)

			if token.AccessToken != test.token.AccessToken {
				t.Errorf("%s: Expected %s but got %s", test.description, test.token.AccessToken, token.AccessToken)
			}

			if token.RefreshToken != test.token.RefreshToken {
				t.Errorf("%s: Expected %s but got %s", test.description, test.token.RefreshToken, token.RefreshToken)
			}

			if token.TokenType != test.token.TokenType {
				t.Errorf("T%s: Expected %s but got %s", test.description, test.token.TokenType, token.TokenType)
			}

			if token.Expiry.Second() > 1800 {
				t.Errorf("%s: Expected expiry time to be less than 1800 but got %d", test.description, token.Expiry.Second())
			}

			if token.Expiry.Second() != test.token.Expiry.Second() {
				t.Errorf("%s: Expected expiry time to be %d but got %d", test.description, test.token.Expiry.Second(), token.Expiry.Second())
			}

		})

		t.Cleanup(func() {
			err := cleanUp("../tests/data/")
			if err != nil {
				t.Errorf("Failed to clean up: %v", err)
			}
		})

	}

}

func TestVerfyToken(t *testing.T) {

	tests := []struct {
		id               int
		content          interface{}
		errorExpected    bool
		createFileBefore bool
		valid            bool
		filePath         string
	}{
		// Valid token
		{
			0,
			oauth2.Token{
				AccessToken:  "askjdsajklsdlkjfasdk",
				RefreshToken: "4ssdjfdsjokdfsopfdopkdfsjksdjiujiosaoi",
				TokenType:    "Bearer",
				Expiry:       time.Now().Add(30 * time.Minute),
			},
			false,
			true,
			true,
			"../tests/data/token.json",
		},
		// Invalid token
		{
			0,
			oauth2.Token{
				AccessToken:  "askjdsajklsdlkjfasdk",
				RefreshToken: "4ssdjfdsjokdfsopfdopkdfsjksdjiujiosaoi",
				TokenType:    "Bearer",
				// Expired 5 minutes ago
				Expiry: time.Now().Add(-5 * time.Minute),
			},
			false,
			true,
			false,
			"../tests/data/token.json",
		},
		// No token file available
		{
			0,
			oauth2.Token{},
			true,
			false,
			false,
			"../tests/data/token.json",
		},
		// Invalid file content
		{
			0,
			"invalid",
			true,
			true,
			false,
			"../tests/data/token.json",
		},
	}

	for index, test := range tests {
		test.id = index + 1

		filePath := filepath.Join("../tests/data/", "token.json")

		if test.createFileBefore {

			err := os.MkdirAll("../tests/data/", 0755)
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create directory: %v", test.id, err)
			}

			file, err := os.Create(filePath)
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create file: %v", test.id, err)
			}

			value, err := json.MarshalIndent(test.content, " ", " ")
			if err != nil {
				t.Errorf("Test Case - %d: Failed to marshal token: %v", test.id, err)
			}

			_, err = file.WriteString(string(value))
			if err != nil {
				t.Errorf("Test Case - %d: Failed to write to file: %v", test.id, err)
			}

			err = file.Close()
			if err != nil {
				t.Errorf("Test Case - %d: Failed to close file: %v", test.id, err)
			}

		}

		valid, _, err := VerfyToken(test.filePath)

		if err != nil && !test.errorExpected {
			t.Errorf("Test Case - %d: Failed to verify token: %v", test.id, err)
		}

		if err == nil && test.errorExpected {
			t.Errorf("Test Case - %d: Expected an error but got none", test.id)
		}

		if valid && test.errorExpected {
			t.Errorf("Test Case - %d: Token is valid but expected an error", test.id)
		}

		if valid != test.valid {
			t.Errorf("Test Case - %d: Expected token with the following valid value %t but got %t", test.id, test.valid, valid)
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
