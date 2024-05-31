// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

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

	err := ntfy.Publish(client, msg, event)
	if err != nil {
		t.Errorf("Publish() returned an error: %v", err)
	}

}

func TestPublishErrorMissingClient(t *testing.T) {

	ntfy := NewNtfy()
	msg := []byte("test message for error")
	event := "test event 2"

	err := ntfy.Publish(nil, msg, event)
	if err == nil {
		t.Errorf("Publish() did not return an error for missing client")

	}
}
