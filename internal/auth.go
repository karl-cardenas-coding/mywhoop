package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// getEndpoint returns the OAuth2 endpoint for the Whoop API
func getEndpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  "https://api.prod.whoop.com/oauth/oauth2/auth",
		TokenURL: "https://api.prod.whoop.com/oauth/oauth2/token",
	}
}

// RefreshToken refreshes the access token
func RefreshToken(ctx context.Context, accessToken, refreshToken string, client *http.Client) (oauth2.Token, error) {

	const (
		url    string = "https://api.prod.whoop.com/oauth/oauth2/token"
		method string = "POST"
	)

	clientID := os.Getenv("WHOOP_CLIENT_ID")
	clientSecret := os.Getenv("WHOOP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return oauth2.Token{}, fmt.Errorf("ClientID and ClientSecret environment variables not set")

	}

	var payloadString strings.Builder
	fmt.Fprintf(&payloadString, "grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s", refreshToken, clientID, clientSecret)
	payloadString.WriteString("&scope=offline read:profile read:recovery read:cycles read:workout read:sleep read:body_measurement")

	payload := strings.NewReader(payloadString.String())

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return oauth2.Token{}, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	res, err := client.Do(req)
	if err != nil {
		return oauth2.Token{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return oauth2.Token{}, err
	}

	var auth AuthCredentials
	err = json.Unmarshal(body, &auth)
	if err != nil {
		LogError(err)
		return oauth2.Token{}, err
	}

	token := oauth2.Token{
		AccessToken:  auth.AccessToken,
		TokenType:    auth.TokenType,
		RefreshToken: auth.RefreshToken,
		Expiry:       time.Now().Local().Add(time.Second * time.Duration(auth.ExpiresIn)),
	}

	return token, nil

}

// GetToken gets the access token from
func GetToken(tokenFilePath string) (string, error) {

	// Set accessToken variable
	accessToken := ""
	clientID := os.Getenv("WHOOP_CLIENT_ID")
	clientSecret := os.Getenv("WHOOP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("ClientID and ClientSecret environment variables not set")

	}

	// Set token file path default
	if tokenFilePath == "" {
		tokenFilePath = "token.json"
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/oauth/redirect",
		Scopes: []string{
			"offline",
			"read:recovery",
			"read:cycles",
			"read:workout",
			"read:sleep",
			"read:profile",
			"read:body_measurement",
		},
		Endpoint: getEndpoint(),
	}

	// Check if token.json file exists
	if _, err := os.Stat("token.json"); err == nil {

		localToken, err := ReadLocalToken(tokenFilePath)
		if err != nil {
			slog.Info("Error reading local token: %v", err)
			return "", err
		}

		if !localToken.Valid() {

			fmt.Println("Local token expired at " + localToken.Expiry.String() + " , refreshing...")

			form := url.Values{}
			form.Add("grant_type", "refresh_token")
			form.Add("refresh_token", localToken.RefreshToken)
			form.Add("client_id", clientID)
			form.Add("client_secret", clientSecret)
			form.Add("scope", "offline")

			body := strings.NewReader(form.Encode())
			req, err := http.NewRequest("POST", AuthURL, body)
			if err != nil {
				return "", err
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return "", err
			}

			// Decode JSON
			var tokenResponse AuthCredentials
			err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
			if err != nil {
				log.Fatal(err)
			}

			// Marshal JSON
			newToken := &oauth2.Token{
				AccessToken:  tokenResponse.AccessToken,
				TokenType:    tokenResponse.TokenType,
				RefreshToken: tokenResponse.RefreshToken,
				Expiry:       time.Now().Local().Add(time.Second * time.Duration(tokenResponse.ExpiresIn)),
			}

			// Write token to file
			writeLocalToken(newToken)

			accessToken = tokenResponse.AccessToken

		} else {

			// Token is valid, use it without refresh
			fmt.Println("Local token valid till " + localToken.Expiry.String() + ", reused without refresh")
			accessToken = localToken.AccessToken

		}

	} else {

		// If token.json not present, start browser authentication flow
		fmt.Println("No token.json found, starting OAuth2 flow")

		// Redirect user to consent page to ask for permission
		authUrl := config.AuthCodeURL("stateidentifier", oauth2.AccessTypeOffline)
		fmt.Println("Visit the URL for the auth dialog: \n\n" + authUrl + "\n")
		fmt.Println("Enter the response URL: ")

		// Wait for user to paste in the response URL
		var respUrl string
		if _, err := fmt.Scan(&respUrl); err != nil {
			return "", err
		}

		// Get response code from response URL string
		parseUrl, _ := url.Parse(respUrl)
		code := parseUrl.Query().Get("code")

		// Exchange response code for token
		accessToken, err := config.Exchange(context.Background(), code)
		if err != nil {
			return "", err
		}

		// Write token to file
		writeLocalToken(accessToken)

	}

	// Return access token and newline
	fmt.Println("")
	return accessToken, nil

}

// Is Token Valid checks if the token is valid
func IsTokenValid(token *oauth2.Token) bool {
	return token.Valid()

}

func ReadLocalToken(filePath string) (oauth2.Token, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return oauth2.Token{}, err
	}
	defer f.Close()

	var token oauth2.Token
	json.NewDecoder(f).Decode(&token)

	return token, nil
}

func writeLocalToken(token *oauth2.Token) {

	f, err := os.Create("token.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	json, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(string(json))

}

type AuthCredentials struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}
