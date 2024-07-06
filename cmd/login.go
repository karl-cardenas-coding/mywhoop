// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
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
	// port is the port to listen on. Default is 8080.
	port string
)

func init() {
	loginCmd.PersistentFlags().BoolVarP(&noAutoOpenBrowser, "no-auto-open", "n", false, "Do not automatically open the browser to authenticate with the Whoop API. ")
	loginCmd.PersistentFlags().StringVarP(&redirectURL, "redirect-url", "r", "/redirect", "The URL path to redirect to after authenticating with the Whoop API. Default is path is /redirect.")
	loginCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "The port to listen on. Default is 8080.")
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

	staticAssets, err := getStaticAssets(GlobalStaticAssets, "web/static")
	if err != nil {
		return err
	}

	config := &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  "http://localhost:" + port + redirectURL,
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
	slog.Debug("Redirect Config", "URL:", "http://localhost:"+port+redirectURL)
	authUrl := internal.GetAuthURL(*config)

	if authUrl == "" {
		return errors.New("unable to get authentication URL. Please check the client ID and client secret are correct")
	}

	// Serve static files from the web/static directory at /static/
	fs := http.FileServer(http.FS(staticAssets))
	// Strip the /static/ prefix from the URL path so that the files are served from / instead of /static/
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", landingPageHandler(GlobalStaticAssets, "web/index.html", authUrl))
	http.HandleFunc("/close", closeHandler)
	http.HandleFunc("/redirect", redirectHandler(GlobalStaticAssets, "web/redirect.html", "web/error.html", config, cliCfg.Credentials.CredentialsFile))

	slog.Info("Listening on port 8080. Visit http://localhost:8080 to autenticate with the Whoop API and get an access token.")
	err = openBrowser("http://localhost:"+port, noAutoOpenBrowser)
	if err != nil {
		slog.Error("unable to open web browser automaticaly", "error", err)
	}
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		return err
	}

	return nil
}

// landingPageHandler handles the landing page and writes the authentication URL to the page
func landingPageHandler(assets fs.FS, indexFile string, authUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp, err := template.ParseFS(assets, indexFile)
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
}

// redirectHandler handles the redirect URL after authenticating with the Whoop API
// and writes the access token to a file
func redirectHandler(assets fs.FS, page, errorPage string, authConf *oauth2.Config, credentialsFilePath string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		slog.Debug("Code received", "code", code)

		if code == "" {
			// slog.Info("no code received.", "Error response status: ", r.Response.StatusCode)
			err := sendErrorTemplate(w, "No authorization code returned by the Whoop authorization server.", http.StatusInternalServerError, errorPage, assets)
			if err != nil {
				slog.Error("unable to send error template", "error", err)
			}
			return
		}

		// Exchange response code for token
		accessToken, err := internal.GetAccessToken(*authConf, code)
		if err != nil {
			slog.Info("unable to get access token", "error", err)
			err := sendErrorTemplate(w, err.Error(), http.StatusInternalServerError, errorPage, assets)
			if err != nil {
				slog.Error("unable to send error template", "error", err)
			}
			return
		}

		err = internal.WriteLocalToken(credentialsFilePath, accessToken)
		if err != nil {
			slog.Debug("Credentials file path", "path", credentialsFilePath)
			slog.Error("unable to write token to file", "error", err)
			err := sendErrorTemplate(w, err.Error(), http.StatusInternalServerError, errorPage, assets)
			if err != nil {
				slog.Error("unable to send error template", "error", err)
			}
			return
		}

		data := PageData{
			CredentialsFilePath: credentialsFilePath,
		}

		tmp, err := template.ParseFS(assets, page)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			slog.Error("unable to parse template", "error", err)
		}

		tmpl := template.Must(tmp, err)
		err = tmpl.Execute(w, data)
		if err != nil {
			slog.Error("unable to execute redirect template", "error", err)
		}
		slog.Info("💾 Access token file created", "path", credentialsFilePath)

	}

}

// getStaticAssets returns the static assets from the embed.FS
func getStaticAssets(f embed.FS, filePath string) (fs.FS, error) {
	return fs.Sub(f, filePath)
}

// closeHandler closes the application after 2 seconds
func closeHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	_, err := w.Write([]byte("Closing application..."))
	if err != nil {
		slog.Error("unable to write response", "error", err)
	}
	defer os.Exit(0)
}

// openBrowser opens a browser to the specified URL
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
func getErrorTemplate(assets fs.FS, file string) (*template.Template, error) {

	t, err := template.ParseFS(assets, file)
	if err != nil {
		return nil, err
	}

	tmpl := template.Must(t, err)
	return tmpl, nil

}

// sendErrorTemplate sends an error message to a response writer
func sendErrorTemplate(w http.ResponseWriter, msg string, statusCode int, file string, assets fs.FS) error {

	tmp, err := getErrorTemplate(assets, file)
	if err != nil {
		return err
	}

	data := PageData{
		Error:      msg,
		StatusCode: statusCode,
	}

	err = tmp.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}
