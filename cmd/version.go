// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

const (
	url = "https://api.github.com/repos/karl-cardenas-coding/mywhoop/releases/latest"
	// IssueMSG is a default message to pass to the user
	IssueMSG = " Please open up a GitHub issue to report this error. https://github.com/karl-cardenas-coding/mywhoop"
)

func init() {
	rootCmd.AddCommand(VersionCmd)
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version number of go-lambda-cleanup",
	Long:  `Prints the current version number of go-lambda-cleanup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version := fmt.Sprintf("mywhoop %s", VersionString)
		slog.Info(version)
		_, message, err := checkForNewRelease(GlobalHTTPClient, VersionString, UserAgent, url)
		if err != nil {
			slog.Error("Error checking for new release", err)
			return err
		}
		slog.Info(message)
		return err
	},
}

func checkForNewRelease(client *http.Client, currentVersion, useragent, url string) (bool, string, error) {
	var (
		output  bool
		message string
		release release
	)

	slog.Info("Checking for new releases")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Error creating new HTTP request", err)
		return output, message, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", useragent)
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error initaiting connection to", url, err)
		return output, message, err
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			msg := fmt.Sprintf("Error status code from Github - %d", resp.StatusCode)
			slog.Error(msg)
			return output, message, fmt.Errorf("error connecting to %s", url)
		}
		// Unmarshal the JSON to the Github Release strcut
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			slog.Error("Error decoding JSON", err)
			return output, message, err
		}

		cVersion, err := version.NewVersion(currentVersion)
		if err != nil {
			slog.Error("Error creating current version", err)
			return output, message, err
		}

		latestVersion, err := version.NewVersion(release.TagName[1:])
		if err != nil {
			slog.Error("Error creating new version", err)
			return output, message, err
		}

		switch cVersion.Compare(latestVersion) {
		case -1:
			message = fmt.Sprintf("There is a new release available: %s \n Download it here - %s", release.TagName, release.HTMLURL)
			output = true
		case 0:
			message = "No new version available"
			output = true
		case 1:
			message = "You are running a pre-release version"
			output = true
		default:
			return output, message, fmt.Errorf("error comparing versions")
		}
	} else {
		return output, message, fmt.Errorf("error connecting to %s", url)
	}

	return output, message, err
}

type release struct {
	URL             string    `json:"url"`
	AssetsURL       string    `json:"assets_url"`
	UploadURL       string    `json:"upload_url"`
	HTMLURL         string    `json:"html_url"`
	ID              int       `json:"id"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Author          author    `json:"author"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []assets  `json:"assets"`
	TarballURL      string    `json:"tarball_url"`
	ZipballURL      string    `json:"zipball_url"`
	Body            string    `json:"body"`
}

type author struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}
type uploader struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}
type assets struct {
	URL                string    `json:"url"`
	ID                 int       `json:"id"`
	NodeID             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	Uploader           uploader  `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}
