package aws

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/internal/utils"
	"golang.org/x/sync/errgroup"
)

type SSOConfig struct {
	StartURL    string
	SSORegion   string
	CLIRegion   string
	SessionName string
}

type ProfileConfig struct {
	Name        string
	AccountID   string
	RoleName    string
	Region      string
	SessionName string
	Output      string
}

type SSOCache struct {
	AccessToken           string `json:"accessToken"`
	ClientID              string `json:"clientId"`
	ClientSecret          string `json:"clientSecret"`
	ExpiresAt             string `json:"expiresAt"`
	RefreshToken          string `json:"refreshToken"`
	Region                string `json:"region"`
	RegistrationExpiresAt string `json:"registrationExpiresAt"`
	StartUrl              string `json:"startUrl"`
}

const configTemplate = `{{ range .Profiles }}[profile {{ .Name }}]
sso_session = {{ .SessionName }}
sso_account_id = {{ .AccountID }}
sso_role_name = {{ .RoleName }}
region = {{ .Region }}
output = {{ .Output }}

{{ end }}
[sso-session {{ .Config.SessionName }}]
sso_start_url = {{ .Config.StartURL }}
sso_region = {{ .Config.SSORegion }}
sso_registration_scopes = sso:account:access

`

type ConfigData struct {
	Profiles []ProfileConfig
	Config   SSOConfig
}

func ConfigureSSO(ssoConfig SSOConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configFilePath := filepath.Join(home, ".aws", "config")
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(ssoConfig.SSORegion),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(5),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	oidcClient := ssooidc.NewFromConfig(cfg)

	regResp, err := registerOIDCClient(oidcClient)
	if err != nil {
		return err
	}
	deviceAuth, err := startDeviceAuth(oidcClient, regResp, ssoConfig.StartURL)
	if err != nil {
		return err
	}
	u.LineBreak()
	u.DeviceCodeFlow(aws.ToString(deviceAuth.VerificationUriComplete), "")
	tokenResp, err := getAccessToken(oidcClient, regResp, deviceAuth.DeviceCode)
	if err != nil {
		return err
	}
	if err := createCacheFile(home, ssoConfig.SessionName, ssoConfig.StartURL, ssoConfig.SSORegion, regResp, tokenResp); err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	log.Debug().Str("package", "aws").Msg("SSO login successful")
	accounts, err := listAccounts(sso.NewFromConfig(cfg), tokenResp.AccessToken)
	if err != nil {
		return err
	}
	log.Debug().Str("package", "aws").Int("accounts", len(accounts.AccountList)).Msg("found accounts")

	configData, err := processAccounts(sso.NewFromConfig(cfg), tokenResp.AccessToken, accounts, ssoConfig)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := writeConfigFile(configFilePath, configData); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	log.Debug().Str("package", "aws").Str("path", configFilePath).Msg("AWS config updated successfully")
	return nil
}

func writeConfigFile(path string, data ConfigData) error {
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}

func processAccounts(client *sso.Client, accessToken *string, accounts sso.ListAccountsOutput, config SSOConfig) (ConfigData, error) {
	var configData ConfigData
	configData.Config = config
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	profileList := []string{}
	g := new(errgroup.Group)
	mu := sync.Mutex{}
	for _, account := range accounts.AccountList {
		g.Go(func() error {
			accountID := aws.ToString(account.AccountId)
			accountName := strings.ToLower(re.ReplaceAllString(aws.ToString(account.AccountName), "-"))
			log.Debug().Str("package", "aws").Str("id", accountID).Str("name", accountName).Msg("processing account")
			roles, err := listAccountRoles(client, accessToken, accountID)
			if err != nil {
				u.PrintWarn("failed to list roles", err)
				return nil
			}
			for _, role := range roles.RoleList {
				roleName := aws.ToString(role.RoleName)
				profileName := accountName
				mu.Lock()
				originalProfileName := profileName
				for _, profile := range configData.Profiles {
					if profile.Name == profileName {
						profileName = fmt.Sprintf("%s-%s-%s", profileName, accountID, roleName)
						log.Debug().Str("package", "aws").Str("originalName", originalProfileName).Str("resolvedName", profileName).Msg("profile name conflict resolved")
						break
					}
				}
				profileList = append(profileList, fmt.Sprintf("%s:%s", profileName, accountID))
				configData.Profiles = append(configData.Profiles, ProfileConfig{
					Name:        profileName,
					AccountID:   accountID,
					RoleName:    roleName,
					Region:      config.CLIRegion,
					SessionName: config.SessionName,
					Output:      "json",
				})
				mu.Unlock()
				log.Debug().Str("package", "aws").Str("profile", profileName).Msg("added profile")
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return ConfigData{}, err
	}
	if err := createProfileList(profileList); err != nil {
		u.PrintWarn("failed to create profile string list", err)
	}
	return configData, nil
}

func createProfileList(profileList []string) error {
	profileFile, err := os.Create("profile-list")
	if err != nil {
		return fmt.Errorf("failed to create profile list file: %w", err)
	}
	defer profileFile.Close()
	writer := bufio.NewWriter(profileFile)
	for _, profile := range profileList {
		fmt.Fprintf(writer, "%s\n", profile)
	}
	writer.Flush()
	return nil
}

func registerOIDCClient(client *ssooidc.Client) (*ssooidc.RegisterClientOutput, error) {
	resp, err := client.RegisterClient(context.TODO(), &ssooidc.RegisterClientInput{
		ClientName: aws.String("anbu-sso-client"),
		ClientType: aws.String("public"),
		Scopes:     []string{"sso:account:access"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register OIDC client: %w", err)
	}
	return resp, nil
}

func startDeviceAuth(client *ssooidc.Client, regResp *ssooidc.RegisterClientOutput, startURL string) (*ssooidc.StartDeviceAuthorizationOutput, error) {
	resp, err := client.StartDeviceAuthorization(context.TODO(), &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     regResp.ClientId,
		ClientSecret: regResp.ClientSecret,
		StartUrl:     aws.String(startURL),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start device authorization: %w", err)
	}
	return resp, nil
}

func createCacheFile(home string, sessionName string, startURL string, region string, regResp *ssooidc.RegisterClientOutput, tokenResp *ssooidc.CreateTokenOutput) error {
	cacheDir := filepath.Join(home, ".aws", "sso", "cache")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	if tokenResp.AccessToken == nil {
		return fmt.Errorf("missing access token in SSO response")
	}
	if regResp.ClientId == nil {
		return fmt.Errorf("missing client ID in SSO registration response")
	}
	if regResp.ClientSecret == nil {
		return fmt.Errorf("missing client secret in SSO registration response")
	}
	refreshToken := ""
	if tokenResp.RefreshToken != nil {
		refreshToken = *tokenResp.RefreshToken
	}
	h := sha1.New()
	h.Write([]byte(sessionName))
	filename := hex.EncodeToString(h.Sum(nil)) + ".json"
	now := time.Now().UTC()
	cache := SSOCache{
		AccessToken:           *tokenResp.AccessToken,
		ClientID:              *regResp.ClientId,
		ClientSecret:          *regResp.ClientSecret,
		ExpiresAt:             now.Add(time.Hour).Format(time.RFC3339),
		RefreshToken:          refreshToken,
		Region:                region,
		RegistrationExpiresAt: now.Add(24 * time.Hour).Format(time.RFC3339),
		StartUrl:              startURL,
	}
	cacheData, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, filename), cacheData, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	log.Debug().Str("package", "aws").Str("file", filename).Msg("cache file created")
	return nil
}

func getAccessToken(client *ssooidc.Client, regResp *ssooidc.RegisterClientOutput, deviceCode *string) (*ssooidc.CreateTokenOutput, error) {
	resp, err := client.CreateToken(context.TODO(), &ssooidc.CreateTokenInput{
		ClientId:     regResp.ClientId,
		ClientSecret: regResp.ClientSecret,
		DeviceCode:   deviceCode,
		GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	return resp, nil
}

func listAccounts(client *sso.Client, accessToken *string) (sso.ListAccountsOutput, error) {
	var accounts sso.ListAccountsOutput
	var nextToken *string
	for {
		resp, err := client.ListAccounts(context.TODO(), &sso.ListAccountsInput{
			AccessToken: accessToken,
			MaxResults:  aws.Int32(100),
			NextToken:   nextToken,
		})
		if err != nil {
			return sso.ListAccountsOutput{}, fmt.Errorf("failed to list accounts: %w", err)
		}
		accounts.AccountList = append(accounts.AccountList, resp.AccountList...)
		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}
		nextToken = resp.NextToken
	}
	return accounts, nil
}

func listAccountRoles(client *sso.Client, accessToken *string, accountID string) (sso.ListAccountRolesOutput, error) {
	var roles sso.ListAccountRolesOutput
	var nextToken *string
	for {
		resp, err := client.ListAccountRoles(context.TODO(), &sso.ListAccountRolesInput{
			AccessToken: accessToken,
			AccountId:   aws.String(accountID),
			MaxResults:  aws.Int32(100),
			NextToken:   nextToken,
		})
		if err != nil {
			return sso.ListAccountRolesOutput{}, fmt.Errorf("failed to list roles for account %s: %w", accountID, err)
		}
		roles.RoleList = append(roles.RoleList, resp.RoleList...)
		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}
		nextToken = resp.NextToken
	}
	return roles, nil
}
