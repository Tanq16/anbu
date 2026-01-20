package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"gopkg.in/ini.v1"
)

type SamlDirectLoginConfig struct {
	Profile      string
	RoleArn      string
	PrincipalArn string
	CLIRegion    string
}

func LoginWithSAMLResponse(config SamlDirectLoginConfig, samlResponseFile string) error {
	var samlAssertion string
	var err error

	if samlResponseFile != "" {
		data, err := os.ReadFile(samlResponseFile)
		if err != nil {
			return fmt.Errorf("failed to read SAML response file: %w", err)
		}
		samlAssertion = strings.TrimSpace(string(data))
	} else {
		samlAssertion = u.InputWithClear("Enter SAML assertion: ")
	}

	if samlAssertion == "" {
		return fmt.Errorf("SAML assertion cannot be empty")
	}

	log.Debug().Msg("authenticating with SAML assertion")
	region := config.CLIRegion
	if region == "" {
		region = "us-east-1"
	}
	cfg, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithRegion(region),
		awsConfig.WithRetryMode(aws.RetryModeAdaptive),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	stsClient := sts.NewFromConfig(cfg)
	duration := int32(3600)
	assumeRoleInput := &sts.AssumeRoleWithSAMLInput{
		RoleArn:         aws.String(config.RoleArn),
		PrincipalArn:    aws.String(config.PrincipalArn),
		SAMLAssertion:   aws.String(samlAssertion),
		DurationSeconds: aws.Int32(duration),
	}
	result, err := stsClient.AssumeRoleWithSAML(context.TODO(), assumeRoleInput)
	if err != nil {
		return fmt.Errorf("failed to assume role with SAML: %w", err)
	}
	log.Debug().Msg("successfully assumed role")

	if err := writeCredentialsToProfile(config.Profile, result.Credentials); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}
	return nil
}

func writeCredentialsToProfile(profile string, credentials *types.Credentials) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	credentialsPath := filepath.Join(home, ".aws", "credentials")
	if err := os.MkdirAll(filepath.Dir(credentialsPath), 0700); err != nil {
		return fmt.Errorf("failed to create .aws directory: %w", err)
	}
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowBooleanKeys:    false,
		IgnoreInlineComment: true,
	}, credentialsPath)
	if err != nil {
		cfg = ini.Empty()
	}
	section, err := cfg.GetSection(profile)
	if err != nil {
		section, err = cfg.NewSection(profile)
		if err != nil {
			return fmt.Errorf("failed to create profile section: %w", err)
		}
	}
	section.Key("aws_access_key_id").SetValue(*credentials.AccessKeyId)
	section.Key("aws_secret_access_key").SetValue(*credentials.SecretAccessKey)
	section.Key("aws_session_token").SetValue(*credentials.SessionToken)
	if err := cfg.SaveTo(credentialsPath); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}
	if err := os.Chmod(credentialsPath, 0600); err != nil {
		return fmt.Errorf("failed to set credentials file permissions: %w", err)
	}
	log.Debug().Str("profile", profile).Msg("credentials written to profile")
	return nil
}
