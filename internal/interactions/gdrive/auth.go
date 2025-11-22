package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func GetDriveService(credentialsFile string) (*drive.Service, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file: %v", err)
	}
	token, err := getOAuthToken(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get OAuth token: %v", err)
	}
	client := config.Client(ctx, token)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}
	return srv, nil
}

func getOAuthToken(config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile, err := getTokenFilePath()
	if err != nil {
		return nil, err
	}
	token, err := tokenFromFile(tokenFile)
	if err == nil {
		if token.Valid() {
			log.Debug().Msgf("existing token retrieved and valid")
			return token, nil
		}
		if token.RefreshToken != "" {
			log.Debug().Msgf("refreshing expired token")
			tokenSource := config.TokenSource(context.Background(), token)
			newToken, err := tokenSource.Token()
			if err != nil {
				log.Debug().Err(err).Msg("token refresh failed, will fall back to new OAuth flow")
				return nil, fmt.Errorf("unable to refresh token: %v", err)
			}
			token = newToken
			if err := saveToken(tokenFile, token); err != nil {
				log.Warn().Msgf("unable to save refreshed token: %v", err)
			}
			return token, nil
		}
	}
	log.Debug().Msgf("no valid token, starting new OAuth flow")
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("\nVisit this URL to authorize Anbu:\n\n%s\n", u.FInfo(authURL))
	fmt.Printf("\nAfter authorizing, enter the authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}
	token, err = config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange auth code for token: %v", err)
	}
	if err := saveToken(tokenFile, token); err != nil {
		log.Warn().Msgf("unable to save new token: %v", err)
	}
	fmt.Println(u.FSuccess("\nAuthentication successful. Token saved."))
	return token, nil
}

func getTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, gdriveTokenFile), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveToken(file string, token *oauth2.Token) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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
