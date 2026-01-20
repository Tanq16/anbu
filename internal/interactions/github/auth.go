package github

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"golang.org/x/oauth2"
)

const (
	deviceCodeURL  = "https://github.com/login/device/code"
	accessTokenURL = "https://github.com/login/oauth/access_token"
)

type GitHubCredentials struct {
	ClientID string `json:"client_id"`
}

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type TokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

func GetGitHubClient(credentialsFile string, pat string) (*http.Client, error) {
	ctx := context.Background()
	var token string
	// Use PAT directly if provided
	if pat != "" {
		token = pat
	} else {
		// otherwise, default to OAuth
		b, err := os.ReadFile(credentialsFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read credentials file: %v", err)
		}
		var creds GitHubCredentials
		if err := json.Unmarshal(b, &creds); err != nil {
			return nil, fmt.Errorf("unable to parse credentials file: %v", err)
		}
		if creds.ClientID == "" {
			return nil, fmt.Errorf("credentials file must contain client_id")
		}
		oauthToken, err := getGitHubOAuthToken(creds.ClientID)
		if err != nil {
			return nil, fmt.Errorf("unable to get OAuth token: %v", err)
		}
		token = oauthToken
	}
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	return oauth2.NewClient(ctx, tokenSource), nil
}

func getGitHubOAuthToken(clientID string) (string, error) {
	tokenFile, err := getGitHubTokenFilePath()
	if err != nil {
		return "", err
	}
	token, err := githubTokenFromFile(tokenFile)
	if err == nil && token != "" {
		log.Debug().Str("op", "github/auth").Msg("existing token retrieved")
		return token, nil
	}
	log.Debug().Str("op", "github/auth").Msg("no valid token, starting device flow")
	deviceResp, err := requestDeviceCode(clientID)
	if err != nil {
		return "", fmt.Errorf("unable to request device code: %v", err)
	}
	fmt.Printf("\nVisit this URL to authorize Anbu:\n\n%s\n", u.FInfo(deviceResp.VerificationURI))
	fmt.Printf("\nEnter the code: %s\n\n", u.FInfo(deviceResp.UserCode))
	fmt.Println("Press Enter after you have completed the authorization in your browser...")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	fmt.Println("Checking for authorization...")
	token, err = pollForAccessToken(clientID, deviceResp)
	if err != nil {
		return "", fmt.Errorf("unable to get access token: %v", err)
	}
	if err := saveGitHubToken(tokenFile, token); err != nil {
		u.PrintWarning("unable to save new token")
		log.Debug().Str("op", "github/auth").Msgf("unable to save new token: %v", err)
	}
	u.PrintSuccess("Authentication successful. Token saved.")
	return token, nil
}

func requestDeviceCode(clientID string) (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", "repo read:org")
	resp, err := http.PostForm(deviceCodeURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to obtain device code (status %d): %s", resp.StatusCode, string(body))
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("unable to parse response: %v", err)
	}
	deviceResp := &DeviceCodeResponse{
		DeviceCode:      values.Get("device_code"),
		UserCode:        values.Get("user_code"),
		VerificationURI: values.Get("verification_uri"),
	}
	if deviceResp.DeviceCode == "" || deviceResp.UserCode == "" {
		return nil, fmt.Errorf("missing required fields in device code response")
	}
	if expiresIn := values.Get("expires_in"); expiresIn != "" {
		fmt.Sscanf(expiresIn, "%d", &deviceResp.ExpiresIn)
	}
	if interval := values.Get("interval"); interval != "" {
		fmt.Sscanf(interval, "%d", &deviceResp.Interval)
	}
	if deviceResp.ExpiresIn == 0 {
		deviceResp.ExpiresIn = 900
	}
	if deviceResp.Interval == 0 {
		deviceResp.Interval = 5
	}
	return deviceResp, nil
}

func pollForAccessToken(clientID string, deviceResp *DeviceCodeResponse) (string, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("device_code", deviceResp.DeviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequest("POST", accessTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response: %v", err)
	}
	bodyStr := string(body)
	log.Debug().Str("op", "github/auth").Msgf("token response (status %d): %s", resp.StatusCode, bodyStr)
	contentType := resp.Header.Get("Content-Type")
	isJSON := strings.Contains(contentType, "application/json") || (len(bodyStr) > 0 && (bodyStr[0] == '{' || bodyStr[0] == '['))
	if isJSON {
		var jsonResp struct {
			AccessToken string `json:"access_token"`
			Error       string `json:"error"`
			ErrorDesc   string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &jsonResp); err == nil {
			if jsonResp.Error != "" {
				if jsonResp.Error == "authorization_pending" {
					return "", fmt.Errorf("authorization still pending - please complete the authorization in your browser and try again")
				}
				return "", fmt.Errorf("token error: %s", jsonResp.ErrorDesc)
			}
			if jsonResp.AccessToken != "" {
				log.Debug().Str("op", "github/auth").Msg("access token received successfully")
				return jsonResp.AccessToken, nil
			}
		}
	}
	values, err := url.ParseQuery(bodyStr)
	if err != nil {
		return "", fmt.Errorf("unable to parse response: %v", err)
	}
	if errorVal := values.Get("error"); errorVal != "" {
		if errorVal == "authorization_pending" {
			return "", fmt.Errorf("authorization still pending - please complete the authorization in your browser and try again")
		}
		errorDesc := values.Get("error_description")
		return "", fmt.Errorf("token error: %s", errorDesc)
	}
	if token := values.Get("access_token"); token != "" {
		log.Debug().Str("op", "github/auth").Msg("access token received successfully")
		return token, nil
	}
	return "", fmt.Errorf("no access token in response: %s", bodyStr)
}

func getGitHubTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	anbuDir := filepath.Join(homeDir, ".anbu")
	if err := os.MkdirAll(anbuDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .anbu directory: %w", err)
	}
	return filepath.Join(anbuDir, githubTokenFile), nil
}

func githubTokenFromFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var tokenData struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(f).Decode(&tokenData)
	return tokenData.AccessToken, err
}

func saveGitHubToken(file string, token string) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	tokenData := map[string]string{
		"access_token": token,
	}
	err = json.NewEncoder(f).Encode(tokenData)
	if err != nil {
		return fmt.Errorf("unable to encode token: %v", err)
	}
	return nil
}
