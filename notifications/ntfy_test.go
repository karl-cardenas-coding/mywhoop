package notifications

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetup(t *testing.T) {
	ntfy := Ntfy{}
	err := ntfy.SetUp()
	if err != nil {
		t.Errorf("Error setting up Ntfy service: %v", err)
	}
}

func TestSend(t *testing.T) {

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
