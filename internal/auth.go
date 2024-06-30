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
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// getEndpoint returns the OAuth2 endpoint for the Whoop API
func getEndpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  DEFAULT_AUTHENTICATION_URL,
		TokenURL: DEFAULT_ACCESS_TOKEN_URL,
	}
}

// GetAuthURL returns the URL to authenticate with the Whoop API
func GetAuthURL(auth oauth2.Config) string {

	return auth.AuthCodeURL("stateidentifier", oauth2.AccessTypeOffline)
}

// GetAccessToken exchanges the access code returned from the authorization flow for an access token
func GetAccessToken(auth oauth2.Config, code string) (*oauth2.Token, error) {
	return auth.Exchange(context.Background(), code)
}

// RefreshToken refreshes the access token
func RefreshToken(ctx context.Context, auth AuthRequest) (oauth2.Token, error) {

	const (
		method string = "POST"
	)

	var payloadString strings.Builder
	var token oauth2.Token
	fmt.Fprintf(&payloadString, "grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s", auth.RefreshToken, auth.ClientID, auth.ClientSecret)
	payloadString.WriteString("&scope=offline read:profile read:recovery read:cycles read:workout read:sleep read:body_measurement")

	payload := strings.NewReader(payloadString.String())

	req, err := http.NewRequest(method, auth.TokenURL, payload)
	if err != nil {
		return oauth2.Token{}, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+auth.AuthToken)

	res, err := auth.Client.Do(req)
	if err != nil {
		return oauth2.Token{}, err
	}

	if res == nil {
		return oauth2.Token{}, errors.New("empty response body from HTTP requests")
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return oauth2.Token{}, err
	}

	var newAuth AuthCredentials
	err = json.Unmarshal(body, &newAuth)
	if err != nil {
		LogError(err)
		return oauth2.Token{}, err
	}

	token = oauth2.Token{
		AccessToken:  newAuth.AccessToken,
		TokenType:    newAuth.TokenType,
		RefreshToken: newAuth.RefreshToken,
		Expiry:       time.Now().Local().Add(time.Second * time.Duration(newAuth.ExpiresIn)),
	}

	return token, nil

}

// writeLocalToken creates file containing the Whoop authentication token
func WriteLocalToken(filePath string, token *oauth2.Token) error {

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	json, err := json.MarshalIndent(token, " ", " ")
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
	err = json.NewDecoder(f).Decode(&token)
	if err != nil {
		slog.Error("unable to decode token file", "error", err)
		return oauth2.Token{}, err
	}

	return token, nil
}
