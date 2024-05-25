// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"net/http"
	"testing"
)

func TestCreateHTTPClient(t *testing.T) {
	tests := []struct {
		id            int
		client        *http.Client
		errorExpected bool
	}{
		{1, nil, true},
		{2, CreateHTTPClient(), false},
	}

	for index, test := range tests {
		test.id = index + 1
		if test.client == nil && !test.errorExpected {
			t.Errorf("Test %d: Expected error, got nil", test.id)
		}
	}

}
