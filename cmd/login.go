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

var (
	// noAutoOpenBrowser is a flag to disable the automatic opening of the browser
	noAutoOpenBrowser bool
	// redirectURL is the URL to redirect to after authenticating with the Whoop API. Default is http://localhost:8080/redirect.
	redirectURL string
)

func init() {
	loginCmd.PersistentFlags().BoolVarP(&noAutoOpenBrowser, "no-auto", "n", false, "Do not automatically open the browser to authenticate with the Whoop API. ")
	loginCmd.PersistentFlags().StringVarP(&redirectURL, "redirect-url", "r", "http://localhost:8080/redirect", "The URL to redirect to after authenticating with the Whoop API. Default is http://localhost:8080/redirect.")
	rootCmd.AddCommand(loginCmd)
}

// PageData is the data structure for the HTML template
type PageData struct {
	// AuthURL is the URL to authenticate with the Whoop API
	AuthURL string
	// CredentialsFilePath is the path to the file where the access token is stored
	CredentialsFilePath string
	// Error is the error message
	Error string
	// StatusCode is the HTTP status code
	StatusCode int
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

	slog.Info(redirectURL)

	config := &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  redirectURL,
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

	fs := http.FileServer(http.Dir("web/static"))

	landingPageHandler := func(w http.ResponseWriter, r *http.Request) {
		tmp, err := template.ParseFiles("web/index.html")
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

		if code == "" {
			slog.Error("no code received.", "Error response status: ", r.Response.StatusCode)
		}

		// Exchange response code for token
		accessToken, err := internal.GetAccessToken(*config, code)
		if err != nil {
			slog.Error("unable to get access token", "error", err)
			pg := PageData{
				Error:      err.Error(),
				StatusCode: http.StatusInternalServerError,
			}
			errorTmp, err := getErrorTemplate("web/error.html")
			if errorTmp != nil {
				err = errorTmp.Execute(w, pg)
				if err != nil {
					slog.Error("unable to execute error template", "error", err)
				}
			}
			if err != nil {
				slog.Error("unable to get error template", "error", err)
			}
		}

		if err == nil {

			err = internal.WriteLocalToken(cliCfg.Credentials.CredentialsFile, accessToken)
			if err != nil {
				slog.Error("unable to write token to file", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			data := PageData{
				CredentialsFilePath: cliCfg.Credentials.CredentialsFile,
			}

			tmp, err := template.ParseFiles("web/redirect.html")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				slog.Error("unable to parse template", "error", err)
			}

			tmpl := template.Must(tmp, err)
			err = tmpl.Execute(w, data)
			if err != nil {
				slog.Error("unable to execute redirect template", "error", err)
			}

		}

	}

	closeAppHandler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		_, err = w.Write([]byte("Closing application..."))
		if err != nil {
			slog.Error("unable to write response", "error", err)
		}
		os.Exit(0)

	}

	http.HandleFunc("/", landingPageHandler)
	http.HandleFunc("/close", closeAppHandler)
	http.HandleFunc("/redirect", redirectHandler)
	http.Handle("/static/", http.StripPrefix("/static/", fs))

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

	c := exec.Command(cmd, args...)
	err := c.Start()
	if err != nil {
		return err
	}

	return nil
}

// getErrorTemplate returns an HTML template from a file
func getErrorTemplate(file string) (*template.Template, error) {

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
	}

	t, err := template.ParseFiles(file)
	if err != nil {
		return nil, err
	}

	tmpl := template.Must(t, err)
	return tmpl, nil

}
