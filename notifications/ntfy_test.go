// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewNtfy(t *testing.T) {

	clearEnvVariables()

	ntfy := NewNtfy()
	if ntfy == nil {
		t.Errorf("Expected Ntfy struct, got nil")
	}

	expected := "errors"
	if ntfy != nil && ntfy.Events != expected {
		t.Errorf("Expected %v, got %v", expected, ntfy.Events)
	}

	clearEnvVariables()
}

func TestSetup(t *testing.T) {

	clearEnvVariables()

	ntfy := NewNtfy()
	ntfy.ServerEndpoint = "http://localhost:8080"
	ntfy.SubscriptionID = "1234"
	ntfy.UserName = "test"

	os.Setenv("NOTIFICATION_NTFY_PASSWORD", "password")

	err := ntfy.SetUp()
	if err != nil {
		t.Errorf("Error setting up Ntfy service: %v", err)
	}

	clearEnvVariables()
}

func TestSetupError(t *testing.T) {

	clearEnvVariables()

	ntfy := NewNtfy()
	ntfy.ServerEndpoint = ""
	ntfy.SubscriptionID = "1234"

	err := ntfy.SetUp()
	if err == nil {
		t.Errorf("Expected error due to missing environment variables but got nil")
	}

	clearEnvVariables()

}

func TestSend(t *testing.T) {

	clearEnvVariables()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Method = "POST"
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Notification sent successfully.")
	}))
	defer ts.Close()

	client := &http.Client{}

	ntfy := NewNtfy()
	ntfy.ServerEndpoint = ts.URL

	err := ntfy.Send(client, []byte("test"), ":tada")
	if err != nil {
		t.Errorf("Error sending Ntfy notification: %v", err)
	}

	clearEnvVariables()

}

func TestSendWithError(t *testing.T) {

	clearEnvVariables()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Notification failed.")
	}))
	defer ts.Close()

	client := &http.Client{}

	ntfy := NewNtfy()
	ntfy.ServerEndpoint = ts.URL

	err := ntfy.Send(client, []byte("test"), "error")
	if err != nil {
		t.Errorf("Error sending Ntfy notification: %v", err)
	}

	clearEnvVariables()

}

func TestSendWithMissingClientError(t *testing.T) {

	clearEnvVariables()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Notification failed.")
	}))
	defer ts.Close()

	ntfy := NewNtfy()
	ntfy.ServerEndpoint = ts.URL

	err := ntfy.Send(nil, []byte("test"), "error")
	if err == nil {
		t.Errorf("Expected error due to missing http client but got nil")

	}
	clearEnvVariables()

}

func TestCanSendMsg(t *testing.T) {

	test := []struct {
		configured string
		event      string
		expected   bool
	}{
		{"all", "errors", true},
		{"all", "success", true},
		{"errors", "errors", true},
		{"errors", "success", false},
		{"success", "success", true},
	}

	for _, tc := range test {
		result := canSendMsg(tc.configured, tc.event)
		if result != tc.expected {
			t.Errorf("Expected %v, got %v", tc.expected, result)
		}
	}

}

func clearEnvVariables() {
	os.Unsetenv("NOTIFICATION_NTFY_PASSWORD")
	os.Unsetenv("NOTIFICATION_NTFY_AUTH_TOKEN")
}

func TestRequiredParams(t *testing.T) {

	test := []struct {
		id            int
		ntfy          Ntfy
		errorExpected bool
	}{
		{0,
			Ntfy{
				ServerEndpoint: "http://localhost:8080",
				SubscriptionID: "1234",
				UserName:       "test",
				Password:       "password",
			}, false},
		{0,
			Ntfy{ServerEndpoint: ""},
			true,
		},
		{0,
			Ntfy{
				ServerEndpoint: "http://localhost:8080",
				SubscriptionID: "",
				UserName:       "test",
				Password:       "password",
			}, true},
		{0,
			Ntfy{
				ServerEndpoint: "http://localhost:8080",
				SubscriptionID: "1234",
				UserName:       "",
				Password:       "",
			}, true},
		{0,
			Ntfy{
				ServerEndpoint: "http://localhost:8080",
				SubscriptionID: "1234",
				UserName:       "",
				Password:       "",
				AccessToken:    "1234",
			}, false},
		{0,
			Ntfy{
				ServerEndpoint: "",
				SubscriptionID: "1234",
				UserName:       "",
				Password:       "",
				AccessToken:    "1234",
			}, true},
		{0,
			Ntfy{
				ServerEndpoint: "http://localhost:8080",
				SubscriptionID: "1234",
				UserName:       "",
				Password:       "",
				AccessToken:    "",
			}, true},
		{0,
			Ntfy{
				ServerEndpoint: "http://localhost:8080",
				SubscriptionID: "1234",
				UserName:       "",
				Password:       "",
				AccessToken:    "1234",
			}, true},
		{0,
			Ntfy{
				ServerEndpoint: "http://localhost:8080",
				SubscriptionID: "1234",
				UserName:       "",
				Password:       "",
				AccessToken:    "",
			}, true},
	}

	for index, tc := range test {
		tc.id = index + 1
		err := checkRequiredParams(tc.ntfy)
		if err != nil {
			if !tc.errorExpected {
				t.Errorf("Test %v failed: unexpected error %v", tc.id, err)
			}
		}
	}

}
