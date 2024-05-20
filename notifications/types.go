// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"net/http"
)

// Notification is an interface that defines the methods for a notification service.
// It requires two method functions SetUp and Send.
// Consumers can use the Publish method to send notifications using the notification service.
type Notification interface {
	// SetUp sets up the notification service and returns an error if the setup fails.
	SetUp() error
	// Send sends a notification using the notification service with the provided data and event.
	Send(clinet *http.Client, data []byte, event string) error
}

// Ntfy is a struct that contains the configuration for the Ntfy notification service.
// Visit https://docs.ntfy.sh/ for more information.
type Ntfy struct {
	// AccessToken is the access token for the Ntfy service. Required if the Ntfy service requires authentication using access token. Provide the access token in the environment variable NOTIFICATION_NTFY_AUTH_TOKEN.
	AccessToken string `yaml:"-"`
	// ServerEndpoint is the endpoint for the Ntfy service.
	ServerEndpoint string `yaml:"serverEndpoint"`
	// SubscriptionID is the subscription ID for the Ntfy service.
	SubscriptionID string `yaml:"subscriptionID"`
	// UserName is the username for the Ntfy service. Required if the Ntfy service requires authentication using username and password. Provide the password in the environment variable NOTIFICATION_NTFY_PASSWORD.
	UserName string `yaml:"userName"`
	// Password is the password for the Ntfy service. Required if the Ntfy service requires authentication using username and password. Provide the password in the environment variable NOTIFICATION_NTFY_PASSWORD.
	Password string `yaml:"-"`
	// Events is a list of events that the Ntfy service can send notifications for. Supported events are errors, success, or all. Default is errors.
	Events string `yaml:"events" validate:"oneof=errors success all"`
}
