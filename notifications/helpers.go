package notifications

import (
	"log/slog"
	"net/http"
)

// Publish sends a notification to the user using the specified notification method.
func Publish(client *http.Client, notificationMethod Notification, msg []byte, event string) {

	if notificationMethod == nil {
		// no notification method specified, do nothing
		return
	}

	if client == nil {
		slog.Info("no http client specified for external notification")
		return
	}

	err := notificationMethod.Send(client, msg, event)
	if err != nil {
		slog.Info("unable to send external notification", "error", err)
	}
}
