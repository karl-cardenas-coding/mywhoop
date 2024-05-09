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
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// getEndpoint returns the OAuth2 endpoint for the Whoop API
func getEndpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  "https://api.prod.whoop.com/oauth/oauth2/auth",
		TokenURL: "https://api.prod.whoop.com/oauth/oauth2/token",
	}
}

// RefreshToken refreshes the access token
func RefreshToken(ctx context.Context, accessToken, refreshToken string, client *http.Client) (oauth2.Token, error) {

	const (
		url    string = "https://api.prod.whoop.com/oauth/oauth2/token"
		method string = "POST"
	)

	clientID := os.Getenv("WHOOP_CLIENT_ID")
	clientSecret := os.Getenv("WHOOP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return oauth2.Token{}, fmt.Errorf("ClientID and ClientSecret environment variables not set")

	}

	var payloadString strings.Builder
	fmt.Fprintf(&payloadString, "grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s", refreshToken, clientID, clientSecret)
	payloadString.WriteString("&scope=offline read:profile read:recovery read:cycles read:workout read:sleep read:body_measurement")

	payload := strings.NewReader(payloadString.String())

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return oauth2.Token{}, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	res, err := client.Do(req)
	if err != nil {
		return oauth2.Token{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return oauth2.Token{}, err
	}

	var auth AuthCredentials
	err = json.Unmarshal(body, &auth)
	if err != nil {
		LogError(err)
		return oauth2.Token{}, err
	}

	token := oauth2.Token{
		AccessToken:  auth.AccessToken,
		TokenType:    auth.TokenType,
		RefreshToken: auth.RefreshToken,
		Expiry:       time.Now().Local().Add(time.Second * time.Duration(auth.ExpiresIn)),
	}

	return token, nil

}

// GetToken is a function that triggers an Oauth flow that endusers can use to aquire a Whoop autentication token using their Whoop client and secret ID.
// The function logic is mostly copied from [INSERT REPO] with some minor modifications.
func GetToken(tokenFilePath string, client *http.Client) (string, error) {

	// Set accessToken variable
	accessToken := ""
	clientID := os.Getenv("WHOOP_CLIENT_ID")
	clientSecret := os.Getenv("WHOOP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("ClientID and ClientSecret environment variables not set")

	}

	// Set token file path default
	if tokenFilePath == "" {
		tokenFilePath = DEFAULT_CREDENTIALS_FILE
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/oauth/redirect",
		Scopes: []string{
			"offline",
			"read:recovery",
			"read:cycles",
			"read:workout",
			"read:sleep",
			"read:profile",
			"read:body_measurement",
		},
		Endpoint: getEndpoint(),
	}

	// Check if token.json file exists
	if _, err := os.Stat(tokenFilePath); err == nil {

		_, localToken, err := VerfyToken(tokenFilePath)
		if err != nil {
			slog.Info("Error reading local token: %v", err)
			return "", err
		}

		if !localToken.Valid() {

			fmt.Println("Local token expired at " + localToken.Expiry.String() + " , refreshing...")

			form := url.Values{}
			form.Add("grant_type", "refresh_token")
			form.Add("refresh_token", localToken.RefreshToken)
			form.Add("client_id", clientID)
			form.Add("client_secret", clientSecret)
			form.Add("scope", "offline")

			body := strings.NewReader(form.Encode())
			req, err := http.NewRequest("POST", AuthURL, body)
			if err != nil {
				return "", err
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			resp, err := client.Do(req)
			if err != nil {
				return "", err
			}

			// Decode JSON
			var tokenResponse AuthCredentials
			err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
			if err != nil {
				return "", err
			}

			// Marshal JSON
			newToken := &oauth2.Token{
				AccessToken:  tokenResponse.AccessToken,
				TokenType:    tokenResponse.TokenType,
				RefreshToken: tokenResponse.RefreshToken,
				Expiry:       time.Now().Local().Add(time.Second * time.Duration(tokenResponse.ExpiresIn)),
			}

			// Write token to file
			err = writeLocalToken(tokenFilePath, newToken)
			if err != nil {
				return "", err
			}

			accessToken = tokenResponse.AccessToken

		} else {

			// Token is valid, use it without refresh
			slog.Info("Local token valid till " + localToken.Expiry.String() + ", reused without refresh")
			accessToken = localToken.AccessToken

		}

	} else {

		// If token.json not present, start browser authentication flow
		slog.Info("No token.json found, starting OAuth2 flow")

		// Redirect user to consent page to ask for permission
		authUrl := config.AuthCodeURL("stateidentifier", oauth2.AccessTypeOffline)
		slog.Info("Visit the following URL for the auth dialog: \n\n" + authUrl + "\n")
		slog.Info("Enter the response URL: ")

		// Wait for user to paste in the response URL
		var respUrl string
		if _, err := fmt.Scan(&respUrl); err != nil {
			return "", err
		}

		// Get response code from response URL string
		parseUrl, _ := url.Parse(respUrl)
		code := parseUrl.Query().Get("code")

		// Exchange response code for token
		accessToken, err := config.Exchange(context.Background(), code)
		if err != nil {
			return "", err
		}

		// Write token to file
		err = writeLocalToken(tokenFilePath, accessToken)
		if err != nil {
			return "", err
		}

	}

	return accessToken, nil

}

// writeLocalToken creates file containing the Whoop authentication token
func writeLocalToken(filePath string, token *oauth2.Token) error {

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	json, err := json.Marshal(token)
	if err != nil {
		return err
	}
	_, err = f.WriteString(string(json))
	if err != nil {
		return err
	}

	return nil

}

// VerifyToken validates that the file containing the Whoop autentication token is valid.
func VerfyToken(filePath string) (bool, oauth2.Token, error) {

	// verify the file exists
	_, err := os.Stat(filePath)
	if err != nil {
		slog.Error("Token file does not exist", "error", err)
		return false, oauth2.Token{}, err
	}

	token, err := ReadTokenFromFile(filePath)
	if err != nil {
		slog.Error("unable to read token file", "error", err)
		return false, oauth2.Token{}, err
	}

	if !token.Valid() {
		LogError(errors.New("invalid or expired auth token"))
		return false, oauth2.Token{}, nil
	}

	return true, token, nil
}

// readTokenFromFile reads a token from a file and returns it as an oauth2.Token
func ReadTokenFromFile(filePath string) (oauth2.Token, error) {

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
