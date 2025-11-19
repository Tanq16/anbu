package box

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"net/http"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"golang.org/x/oauth2"
)

func GetBoxClient(credentialsFile string) (*http.Client, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}
	var creds BoxCredentials
	if err := json.Unmarshal(b, &creds); err != nil {
		return nil, fmt.Errorf("unable to parse credentials file: %v", err)
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, fmt.Errorf("credentials file must contain client_id and client_secret")
	}
	config := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://account.box.com/api/oauth2/authorize",
			TokenURL: "https://api.box.com/oauth2/token",
		},
		RedirectURL: redirectURI,
		Scopes:      []string{"root_readwrite"},
	}
	token, err := getBoxOAuthToken(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get OAuth token: %v", err)
	}
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to refresh token: %v", err)
	}
	if newToken.AccessToken != token.AccessToken {
		log.Debug().Str("op", "box/auth").Msg("access token was refreshed")
		saveBoxToken(newToken)
	}
	return oauth2.NewClient(ctx, tokenSource), nil
}

func getBoxOAuthToken(config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile, err := getBoxTokenFilePath()
	if err != nil {
		return nil, err
	}
	token, err := boxTokenFromFile(tokenFile)
	if err == nil {
		if token.Valid() {
			log.Debug().Str("op", "box/auth").Msg("existing token retrieved and valid")
			return token, nil
		}
		if token.RefreshToken != "" {
			log.Debug().Str("op", "box/auth").Msg("refreshing expired token")
			tokenSource := config.TokenSource(context.Background(), token)
			newToken, err := tokenSource.Token()
			if err != nil {
				return nil, fmt.Errorf("unable to refresh token: %v", err)
			}
			token = newToken
			if err := saveBoxToken(token); err != nil {
				log.Warn().Str("op", "box/auth").Msgf("unable to save refreshed token: %v", err)
			}
			return token, nil
		}
	}
	log.Debug().Str("op", "box/auth").Msg("no valid token, starting new OAuth flow")
	state := fmt.Sprintf("st%d", os.Getpid())
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("\nVisit this URL to authorize Anbu:\n\n%s\n", u.FInfo(authURL))
	fmt.Printf("\nAfter authorizing, you will be redirected to a 'localhost' URL.\n")
	fmt.Printf("Copy the *entire* 'localhost' URL from your browser and paste it here: ")
	var redirectURLStr string
	if _, err := fmt.Scanln(&redirectURLStr); err != nil {
		return nil, fmt.Errorf("unable to read redirect URL: %v", err)
	}
	parsedURL, err := url.Parse(redirectURLStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse the pasted URL: %v", err)
	}
	code := parsedURL.Query().Get("code")
	returnedState := parsedURL.Query().Get("state")
	if code == "" {
		return nil, fmt.Errorf("pasted URL did not contain an authorization 'code'")
	}
	if returnedState != state {
		return nil, fmt.Errorf("CSRF state mismatch. Expected '%s' but got '%s'", state, returnedState)
	}
	fmt.Println("Trading code for token...")
	token, err = config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange auth code for token: %v", err)
	}
	if err := saveBoxToken(token); err != nil {
		log.Warn().Str("op", "box/auth").Msgf("unable to save new token: %v", err)
	}
	fmt.Println(u.FSuccess("\nAuthentication successful. Token saved."))
	return token, nil
}

func getBoxTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, boxTokenFile), nil
}

func boxTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveBoxToken(token *oauth2.Token) error {
	tokenFile, err := getBoxTokenFilePath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("unable to encode token: %v", err)
	}
	return nil
}
