package notifications

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublishSuccess(t *testing.T) {

	client := &http.Client{}
	ntfy := NewNtfy()
	msg := []byte("test message")
	event := "test event"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Notification success.")
	}))
	defer ts.Close()

	err := Publish(client, ntfy, msg, event)
	if err != nil {
		t.Errorf("Publish() returned an error: %v", err)
	}

}

func TestPublishErrorMissingClient(t *testing.T) {

	ntfy := NewNtfy()
	msg := []byte("test message for error")
	event := "test event 2"

	err := Publish(nil, ntfy, msg, event)
	if err == nil {
		t.Errorf("Publish() did not return an error for missing client")

	}

}

func TestPublishErrorMissingNotificationMethod(t *testing.T) {

	client := &http.Client{}
	msg := []byte("test message for error")
	event := "test event 3"

	err := Publish(client, nil, msg, event)
	if err == nil {
		t.Errorf("Publish() did not return an error for missing notification method")
	}

}