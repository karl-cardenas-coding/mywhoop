package notifications

import (
	"log/slog"
)

// externalNotificaton sends a notification to the user using the specified notification method.
func EternalNotificaton(notificationMethod Notification, msg []byte, emoji string) {

	if notificationMethod == nil {
		// no notification method specified, do nothing
		return
	}

	err := notificationMethod.Send(msg, emoji)
	if err != nil {
		slog.Info("unable to send external notification", "error", err)
	}
}
