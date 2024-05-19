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
	if ntfy.Events != expected {
		t.Errorf("Expected %v, got %v", expected, ntfy.Events)
	}

	clearEnvVariables()
}

func TestSetup(t *testing.T) {

	clearEnvVariables()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Notifcation sent successfully.")
	}))
	defer ts.Close()

	ntfy := NewNtfy()
	ntfy.ServerEndpoint = ts.URL
	ntfy.SubscriptionID = "1234"
	ntfy.UserName = "test"

	os.Setenv("NOTIFICATION_NTFY_PASSWORD", "password")

	err := ntfy.SetUp()
	if err != nil {
		t.Errorf("Error setting up Ntfy service: %v", err)
	}

	clearEnvVariables()
}

func TestSend(t *testing.T) {

	clearEnvVariables()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Notifcation sent successfully.")
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
