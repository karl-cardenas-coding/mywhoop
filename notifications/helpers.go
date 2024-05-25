// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"errors"
	"net/http"
)

// Publish sends a notification to the user using the specified notification method.
func Publish(client *http.Client, notificationMethod Notification, msg []byte, event string) error {

	if notificationMethod == nil {
		// no notification method specified, do nothing
		return errors.New("no notification method specified")
	}

	if client == nil {
		return errors.New("no http client specified for external notification")
	}

	err := notificationMethod.Send(client, msg, event)
	if err != nil {
		return err
	}

	return nil
}
