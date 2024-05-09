package internal

import (
	"testing"

	"golang.org/x/oauth2"
)

func TestGetEndpoint(t *testing.T) {

	expected := oauth2.Endpoint{
		AuthURL:  "https://api.prod.whoop.com/oauth/oauth2/auth",
		TokenURL: "https://api.prod.whoop.com/oauth/oauth2/token",
	}

	got := getEndpoint()

	if got.AuthURL != expected.AuthURL {
		t.Errorf("an error occured. Expected %s but got %s", expected.AuthURL, got.AuthURL)
	}

	if got.TokenURL != expected.TokenURL {
		t.Errorf("an error occured. Expected %s but got %s", expected.TokenURL, got.TokenURL)
	}

}
