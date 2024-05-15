package notifications

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
)

// SetUp sets up the Ntfy service.
func (n *Ntfy) SetUp() error {

	pwd := os.Getenv("NOTIFICATION_NTFY_PASSWORD")
	token := os.Getenv("NOTIFICATION_NTFY_AUTH_TOKEN")

	n.Password = pwd
	n.AccessToken = token

	err := checkRequiredParams(*n)
	if err != nil {
		return err
	}

	// Set up the Ntfy service
	return nil
}

// Send sends a notification using the Ntfy service with the provided data.
func (n *Ntfy) Send(data []byte) error {

	// convert data to io.Reader
	payload := bytes.NewReader(data)

	req, err := http.NewRequestWithContext(context.Background(), "POST", n.ServerEndpoint+"/"+n.SubscriptionID, payload)
	if err != nil {
		return err
	}

	if n.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+n.AccessToken)
	}

	if n.UserName != "" && n.Password != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(n.UserName + ":" + n.Password))
		req.Header.Set("Authorization", "Basic "+encoded)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if response != nil {

		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			slog.Info("unable to read response body", "error", err)
			return err
		}

		if response.StatusCode != http.StatusOK {
			slog.Info("error sending notification", "status", response.Status, "body", string(body))
			return errors.New("error sending notification")
		}

		slog.Info("notification sent", "status", response.Status)

	}

	return nil
}

// checkRequiredParams checks if the required parameters are provided. If a required parameter is not provided, it returns an error.
func checkRequiredParams(ntfy Ntfy) error {

	if ntfy.ServerEndpoint == "" {
		return errors.New("serverEndpoint is required")
	}
	if ntfy.SubscriptionID == "" {
		return errors.New("subscriptionID is required")
	}

	switch {
	// Check if either username or access token is provided
	case ntfy.UserName != "" && ntfy.Password != "":
		if ntfy.AccessToken != "" {
			return errors.New("provide either username and password or access token")
		}
	// Check if access token is provided, otherwise check if username and password are provided
	case ntfy.AccessToken != "":
		if ntfy.UserName != "" || ntfy.Password != "" {
			return errors.New("provide either username and password or access token")
		}
	default:
		return errors.New("unable to determine if credentials are provided")

	}

	return nil
}
