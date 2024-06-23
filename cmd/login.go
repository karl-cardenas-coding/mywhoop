// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Whoop API and get an access token",
	Long:  "Authenticate with Whoop API and get an access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		return login()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

// login authenticates with Whoop API and gets an access token
func login() error {
	err := InitLogger(&Configuration)
	if err != nil {
		return err
	}

	id := os.Getenv("WHOOP_CLIENT_ID")
	secret := os.Getenv("WHOOP_CLIENT_SECRET")

	if id == "" || secret == "" {
		return errors.New("the required env variables WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET are not set")
	}

	// cfg := Configuration
	auth := internal.AuthRequest{
		ClientID:         id,
		ClientSecret:     secret,
		AuthorizationURL: internal.DEFAULT_AUTHENTICATION_URL,
		TokenURL:         internal.DEFAULT_ACCESS_TOKEN_URL,
	}

	slog.Info("Starting login application helper")
	fs := http.FileServer(http.Dir("html/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	landingPageHandler := func(w http.ResponseWriter, r *http.Request) {
		tmp, err := template.ParseFiles("html/index.html")
		if err != nil {
			slog.Error("unable to parse template", "error", err)
		}
		tmpl := template.Must(tmp, err)
		err = tmpl.Execute(w, auth)
		if err != nil {
			slog.Error("unable to execute template", "error", err)
		}

	}

	submitHandler := func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		slog.Info("Username and password received", "username", username, "password", password)
		rsp, err := w.Write([]byte(`<div class="container">
		<div class="message">
			<p>You have successfully authenticated with the Whoop API 🎉.</p>
			<p>A file was created in the local directory titled <strong>token.json</strong>. Use the button below to close the application.</p> <p> ⚠️ You must manually close this window - Sorry browser security settings 🔐</p>
		</div>
		<button hx-post="/close" hx-trigger="click" class="close-button">Close CLI Application</button>
	</div>`))
		if err != nil {
			slog.Error("unable to write response", "error", err)
		}
		slog.Info("Response written", "response", rsp)

	}

	closeAppHandler := func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Closing login application helper")
		time.Sleep(2 * time.Second)
		w.Write([]byte("Closing application..."))
		os.Exit(0)

	}

	http.HandleFunc("/", landingPageHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/close", closeAppHandler)

	slog.Info("Listening on port 8080. Visit http://localhost:8080 to autenticate with the Whoop API and get an access token.")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return err
	}

	return nil
}
