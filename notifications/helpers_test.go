package notifications

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublish(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Notifcation sent successfully.")
	}))
	defer ts.Close()

	client := &http.Client{}

	ntfy := NewNtfy()

	Publish(client, ntfy, []byte("test"), "errors")

}
