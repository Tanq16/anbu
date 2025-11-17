package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

const (
	awsFedEndpoint = "https://signin.aws.amazon.com/federation"
	consoleBase    = "https://console.aws.amazon.com/"
	defaultIssuer  = "aws-console-tool"
	maxDuration    = 3600
)

var consolePolicy = map[string]any{
	"Version": "2012-10-17",
	"Statement": []map[string]any{
		{
			"Effect":   "Allow",
			"Action":   []string{"*"},
			"Resource": []string{"*"},
		},
	},
}

func GenerateConsoleURLFromProfile(profile string) (string, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion("us-east-1"),
		config.WithRetryMode(aws.RetryModeAdaptive),
	)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("failed to validate credentials: %w", err)
	}
	var credentials *ststypes.Credentials
	if strings.Contains(*identity.Arn, ":assumed-role/") {
		creds, err := cfg.Credentials.Retrieve(context.TODO())
		if err != nil {
			return "", fmt.Errorf("failed to retrieve credentials: %w", err)
		}
		credentials = &ststypes.Credentials{
			AccessKeyId:     aws.String(creds.AccessKeyID),
			SecretAccessKey: aws.String(creds.SecretAccessKey),
			SessionToken:    aws.String(creds.SessionToken),
		}
	} else {
		credentials, err = getFederationToken(stsClient, "console-session", maxDuration)
		if err != nil {
			return "", fmt.Errorf("failed to get federation token: %w", err)
		}
	}
	consoleURL, err := generateConsoleURL(credentials)
	if err != nil {
		return "", fmt.Errorf("failed to generate console URL: %w", err)
	}
	return consoleURL, nil
}

func getFederationToken(stsClient *sts.Client, federationName string, duration int) (*ststypes.Credentials, error) {
	policyBytes, err := json.Marshal(consolePolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy: %w", err)
	}
	result, err := stsClient.GetFederationToken(context.TODO(), &sts.GetFederationTokenInput{
		Name:            aws.String(federationName),
		Policy:          aws.String(string(policyBytes)),
		DurationSeconds: aws.Int32(int32(duration)),
	})
	if err != nil {
		return nil, err
	}
	return result.Credentials, nil
}

func generateConsoleURL(credentials *ststypes.Credentials) (string, error) {
	sessionData := map[string]string{
		"sessionId":    *credentials.AccessKeyId,
		"sessionKey":   *credentials.SecretAccessKey,
		"sessionToken": *credentials.SessionToken,
	}
	sessionDataBytes, err := json.Marshal(sessionData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %w", err)
	}
	federationURL := fmt.Sprintf("%s?Action=getSigninToken&Session=%s", awsFedEndpoint, url.QueryEscape(string(sessionDataBytes)))
	resp, err := http.Get(federationURL)
	if err != nil {
		return "", fmt.Errorf("failed to get sign-in token: %w", err)
	}
	defer resp.Body.Close()
	var tokenResponse struct {
		SigninToken string `json:"SigninToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode sign-in token response: %w", err)
	}
	consoleURL := fmt.Sprintf("%s?Action=login&Issuer=%s&Destination=%s&SigninToken=%s", awsFedEndpoint, defaultIssuer, consoleBase, url.QueryEscape(tokenResponse.SigninToken))
	return consoleURL, nil
}
