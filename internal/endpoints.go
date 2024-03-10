package internal

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// getUserProfile returns the user profile from the Whoop API
func (u User) GetData(ctx context.Context, client *http.Client, authToken string) (*User, error) {

	const (
		url    = "https://api.prod.whoop.com/developer/v1/user/profile/basic"
		method = "GET"
	)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		slog.Error("unable to create request", err)
		return nil, err
	}
	authHeader := "Bearer " + authToken
	req.Header.Add("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		slog.Error("unable to get user data", err)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("unable to read response body", err)
		return nil, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		slog.Error("unable to unmarshal data", err)
		return nil, err
	}

	return &user, nil

}
