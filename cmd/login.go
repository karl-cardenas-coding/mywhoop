// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"text/template"
	"time"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Whoop API and get an access token",
	Long:  "Authenticate with Whoop API and get an access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		return login()
	},
}

var noAutoOpenBrowser bool

func init() {
	loginCmd.PersistentFlags().BoolVarP(&noAutoOpenBrowser, "no-auto", "n", false, "Do not automatically open the browser to authenticate with the Whoop API.")
	rootCmd.AddCommand(loginCmd)
}

// PageData is the data structure for the HTML template
type PageData struct {
	// AuthURL is the URL to authenticate with the Whoop API
	AuthURL string
}

// login authenticates with Whoop API and gets an access token
func login() error {
	err := InitLogger(&Configuration)
	if err != nil {
		return err
	}

	cliCfg := Configuration

	id := os.Getenv("WHOOP_CLIENT_ID")
	secret := os.Getenv("WHOOP_CLIENT_SECRET")

	if id == "" || secret == "" {
		return errors.New("the required env variables WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET are not set")
	}

	config := &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  "http://localhost:8080/redirect",
		Scopes: []string{
			"offline",
			"read:recovery",
			"read:cycles",
			"read:workout",
			"read:sleep",
			"read:profile",
			"read:body_measurement",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  internal.DEFAULT_AUTHENTICATION_URL,
			TokenURL: internal.DEFAULT_ACCESS_TOKEN_URL,
		},
	}

	authUrl := internal.GetAuthURL(*config)

	slog.Info("Starting login application helper")
	fs := http.FileServer(http.Dir("html/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	landingPageHandler := func(w http.ResponseWriter, r *http.Request) {
		tmp, err := template.ParseFiles("html/index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			slog.Error("unable to parse template", "error", err)
		}

		data := PageData{
			AuthURL: authUrl,
		}

		tmpl := template.Must(tmp, err)
		err = tmpl.Execute(w, data)
		if err != nil {
			slog.Error("unable to execute template", "error", err)
		}

	}

	redirectHandler := func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		slog.Debug("Code received", "code", code)

		// Exchange response code for token
		accessToken, err := internal.GetAccessToken(*config, code)
		if err != nil {
			slog.Error("unable to get access token", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		err = internal.WriteLocalToken(cliCfg.Credentials.CredentialsFile, accessToken)
		if err != nil {
			slog.Error("unable to write token to file", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		_, err = w.Write([]byte(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta http-equiv="X-UA-Compatible" content="IE=edge">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>MyWhoop Token Helper</title>
				<script src="/static/dependencies/htmx.min.js"></script>
				<link rel="stylesheet" href="/static/styles/styles.css">
				<link rel="apple-touch-icon" sizes="180x180" href="/static/apple-touch-icon.png">
				<link rel="icon" type="image/png" sizes="32x32" href="/static/favicon-32x32.png">
				<link rel="icon" type="image/png" sizes="16x16" href="/static/favicon-16x16.png">
				<link rel="manifest" href="/static/site.webmanifest">
			</head>
				<body>
			<div class="container">
					<div class="message">
						<p>You have successfully authenticated with the Whoop API 🎉.</p>
						<p>A file was created in the specified credentials file path titled <strong>token.json</strong>. Use the button below to close the application.</p> <p> ⚠️ You must manually close this window - Sorry browser security settings 🔐</p>
					</div>
					<button hx-post="/close" hx-trigger="click" class="close-button">Close CLI Application</button>
				</div>
			</body>
			</html>`))
		if err != nil {
			slog.Error("unable to write response", "error", err)
		}
	}

	closeAppHandler := func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Closing login application helper")
		time.Sleep(2 * time.Second)
		w.Write([]byte("Closing application..."))
		os.Exit(0)

	}

	http.HandleFunc("/", landingPageHandler)
	http.HandleFunc("/close", closeAppHandler)
	http.HandleFunc("/redirect", redirectHandler)

	slog.Info("Listening on port 8080. Visit http://localhost:8080 to autenticate with the Whoop API and get an access token.")
	err = openBrowser("http://localhost:8080", noAutoOpenBrowser)
	if err != nil {
		return err
	}
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return err
	}

	return nil
}

func openBrowser(url string, disableCmd bool) error {
	var cmd string
	var args []string

	if disableCmd {
		return nil
	}

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}
