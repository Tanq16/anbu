package cloudCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuCloud "github.com/tanq16/anbu/internal/cloud/aws"
	u "github.com/tanq16/anbu/utils"
)

var awsIidcLoginFlags struct {
	startURL    string
	ssoRegion   string
	cliRegion   string
	sessionName string
}

var awsSamlDirectLoginFlags struct {
	samlResponseFile string
	cliRegion        string
	roleArn          string
	principalArn     string
	profile          string
}

var awsCliUiFlags struct {
	profile string
}

var AwsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Helper utilities for AWS",
}

var awsIidcLoginCmd = &cobra.Command{
	Use:   "iidc-login",
	Short: "Configure AWS SSO with IAM Identity Center for multi-role access",
	Run: func(cmd *cobra.Command, args []string) {
		if awsIidcLoginFlags.startURL == "" || awsIidcLoginFlags.ssoRegion == "" {
			u.PrintFatal("Both --start-url and --sso-region flags are required", nil)
		}
		config := anbuCloud.SSOConfig{
			StartURL:    awsIidcLoginFlags.startURL,
			SSORegion:   awsIidcLoginFlags.ssoRegion,
			CLIRegion:   awsIidcLoginFlags.cliRegion,
			SessionName: awsIidcLoginFlags.sessionName,
		}
		if err := anbuCloud.ConfigureSSO(config); err != nil {
			u.PrintFatal("Failed to configure SSO", err)
		}
		u.PrintSuccess("Successfully configured AWS SSO")
	},
}

var awsSamlDirectLoginCmd = &cobra.Command{
	Use:   "saml-direct-login",
	Short: "Login to AWS CLI with SAML response grabbed from a browser session directly",
	Run: func(cmd *cobra.Command, args []string) {
		if awsSamlDirectLoginFlags.roleArn == "" || awsSamlDirectLoginFlags.principalArn == "" {
			u.PrintFatal("Both --role-arn and --principal-arn flags are required", nil)
		}
		if awsSamlDirectLoginFlags.profile == "" {
			awsSamlDirectLoginFlags.profile = "default"
		}
		config := anbuCloud.SamlDirectLoginConfig{
			Profile:      awsSamlDirectLoginFlags.profile,
			RoleArn:      awsSamlDirectLoginFlags.roleArn,
			PrincipalArn: awsSamlDirectLoginFlags.principalArn,
			CLIRegion:    awsSamlDirectLoginFlags.cliRegion,
		}
		if err := anbuCloud.LoginWithSAMLResponse(config, awsSamlDirectLoginFlags.samlResponseFile); err != nil {
			u.PrintFatal("Failed to login via SAML", err)
		}
		u.PrintSuccess(fmt.Sprintf("Successfully logged in via SAML (profile: %s)", awsSamlDirectLoginFlags.profile))
	},
}

var awsCliUiCmd = &cobra.Command{
	Use:   "cli-ui",
	Short: "Get a console URL from an AWS CLI profile with a pre-signed URL valid for up to 12 hours",
	Run: func(cmd *cobra.Command, args []string) {
		consoleURL, err := anbuCloud.GenerateConsoleURLFromProfile(awsCliUiFlags.profile)
		if err != nil {
			u.PrintFatal("Failed to generate console URL", err)
		}
		u.PrintInfo("Console URL:")
		u.PrintGeneric(consoleURL)
		u.PrintInfo("URL valid for 12 hours")
	},
}

func init() {
	AwsCmd.AddCommand(awsIidcLoginCmd)
	AwsCmd.AddCommand(awsSamlDirectLoginCmd)
	AwsCmd.AddCommand(awsCliUiCmd)

	awsIidcLoginCmd.Flags().StringVarP(&awsIidcLoginFlags.startURL, "start-url", "u", "", "AWS SSO start URL (e.g. https://my-sso.awsapps.com/start)")
	awsIidcLoginCmd.Flags().StringVarP(&awsIidcLoginFlags.ssoRegion, "sso-region", "r", "us-east-1", "AWS SSO region (e.g. us-east-1)")
	awsIidcLoginCmd.Flags().StringVarP(&awsIidcLoginFlags.cliRegion, "cli-region", "e", "us-east-1", "Default AWS CLI region for the new profiles")
	awsIidcLoginCmd.Flags().StringVarP(&awsIidcLoginFlags.sessionName, "session-name", "n", "my-sso", "SSO session name to use in the config file")

	awsSamlDirectLoginCmd.Flags().StringVarP(&awsSamlDirectLoginFlags.roleArn, "role-arn", "r", "", "AWS IAM role ARN to assume")
	awsSamlDirectLoginCmd.Flags().StringVarP(&awsSamlDirectLoginFlags.principalArn, "principal-arn", "i", "", "AWS SAML provider ARN")
	awsSamlDirectLoginCmd.Flags().StringVarP(&awsSamlDirectLoginFlags.samlResponseFile, "file", "f", "", "File containing SAML assertion (otherwise reads from stdin)")
	awsSamlDirectLoginCmd.Flags().StringVarP(&awsSamlDirectLoginFlags.profile, "profile", "p", "default", "AWS profile name to write credentials to")
	awsSamlDirectLoginCmd.Flags().StringVarP(&awsSamlDirectLoginFlags.cliRegion, "cli-region", "e", "us-east-1", "Default AWS CLI region for the profile")

	awsCliUiCmd.Flags().StringVarP(&awsCliUiFlags.profile, "profile", "p", "default", "AWS profile to use for console URL generation (default: 'default')")
}
