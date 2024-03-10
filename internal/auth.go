package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

func GetToken() (string, error) {

	// Set accessToken variable
	accessToken := ""
	clientID := os.Getenv("WHOOP_CLIENT_ID")
	clientSecret := os.Getenv("WHOOP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("ClientID and ClientSecret environment variables not set")

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

		localToken := readLocalToken()

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
			var tokenResponse TokenLocalFile
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

func readLocalToken() oauth2.Token {

	f, err := os.Open("token.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var token oauth2.Token
	json.NewDecoder(f).Decode(&token)

	return token
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

type TokenLocalFile struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}
