// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"

	"golang.org/x/oauth2"
)

func TestGetEndpoint(t *testing.T) {

	expected := oauth2.Endpoint{
		AuthURL:  DEFAULT_AUTHENTICATION_URL,
		TokenURL: DEFAULT_ACCESS_TOKEN_URL,
	}

	got := getEndpoint()

	if got.AuthURL != expected.AuthURL {
		t.Errorf("an error occured. Expected %s but got %s", expected.AuthURL, got.AuthURL)
	}

	if got.TokenURL != expected.TokenURL {
		t.Errorf("an error occured. Expected %s but got %s", expected.TokenURL, got.TokenURL)
	}

}
