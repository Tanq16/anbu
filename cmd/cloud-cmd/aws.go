package cloudCmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuCloud "github.com/tanq16/anbu/internal/cloud/aws"
)

var awsIidcLoginFlags struct {
	startURL    string
	ssoRegion   string
	cliRegion   string
	sessionName string
}

var awsCliUiFlags struct {
	profile string
}

var AwsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Helper utilities for AWS",
	Long: `Helper utilities for AWS.

Subcommands:
  iidc-login: Configure AWS SSO with IAM Identity Center for multi-role access
  cli-ui:     Generate AWS console URL from a local CLI profile`,
}

var awsIidcLoginCmd = &cobra.Command{
	Use:   "iidc-login",
	Short: "Configure AWS SSO with IAM Identity Center",
	Run: func(cmd *cobra.Command, args []string) {
		if awsIidcLoginFlags.startURL == "" || awsIidcLoginFlags.ssoRegion == "" {
			log.Fatal().Msg("Both --start-url and --sso-region flags are required")
		}

		config := anbuCloud.SSOConfig{
			StartURL:    awsIidcLoginFlags.startURL,
			SSORegion:   awsIidcLoginFlags.ssoRegion,
			CLIRegion:   awsIidcLoginFlags.cliRegion,
			SessionName: awsIidcLoginFlags.sessionName,
		}

		if err := anbuCloud.ConfigureSSO(config); err != nil {
			log.Fatal().Err(err).Msg("failed to configure SSO")
		}

		log.Info().Msg("Successfully configured AWS SSO")
	},
}

var awsCliUiCmd = &cobra.Command{
	Use:   "cli-ui",
	Short: "Get a console URL from an AWS CLI profile",
	Run: func(cmd *cobra.Command, args []string) {
		consoleURL, err := anbuCloud.GenerateConsoleURLFromProfile(awsCliUiFlags.profile)
		if err != nil {
			log.Fatal().Err(err).Str("profile", awsCliUiFlags.profile).Msg("failed to generate console URL")
		}

		log.Info().Str("profile", awsCliUiFlags.profile).Str("url", consoleURL).Msg("Console URL")
		log.Info().Msg("URL valid for 12 hours")
	},
}

func init() {
	AwsCmd.AddCommand(awsIidcLoginCmd)
	AwsCmd.AddCommand(awsCliUiCmd)

	awsIidcLoginCmd.Flags().StringVarP(&awsIidcLoginFlags.startURL, "start-url", "u", "", "AWS SSO start URL (e.g. https://my-sso.awsapps.com/start)")
	awsIidcLoginCmd.Flags().StringVarP(&awsIidcLoginFlags.ssoRegion, "sso-region", "r", "us-east-1", "AWS SSO region (e.g. us-east-1)")
	awsIidcLoginCmd.Flags().StringVarP(&awsIidcLoginFlags.cliRegion, "cli-region", "e", "us-east-1", "Default AWS CLI region for the new profiles")
	awsIidcLoginCmd.Flags().StringVarP(&awsIidcLoginFlags.sessionName, "session-name", "n", "my-sso", "SSO session name to use in the config file")

	awsCliUiCmd.Flags().StringVarP(&awsCliUiFlags.profile, "profile", "p", "default", "AWS profile to use for console URL generation (default: 'default')")
}
