// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"crypto/tls"
	"net/http"
)

// createHTTPClient creates an HTTP client with TLS
func CreateHTTPClient() *http.Client {

	// Setup client header to use TLS 1.2
	tr := &http.Transport{
		// Reads PROXY configuration from environment variables
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// Needed due to custom client being leveraged, otherwise HTTP2 will not be used.
	tr.ForceAttemptHTTP2 = true

	// Create the client
	return &http.Client{Transport: tr}
}
