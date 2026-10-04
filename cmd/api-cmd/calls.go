package apiCmd

import (
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Read vault secrets",
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets without values",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		call(http.MethodGet, "/api/secrets", nil)
	},
}

var secretsGetCmd = &cobra.Command{
	Use:   "get <ref>",
	Short: "Print a secret, or SSO temporary credentials for <secret>:<profile>",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		call(http.MethodGet, "/api/secrets/"+segment(args[0]), nil)
	},
}

var secretsTOTPCmd = &cobra.Command{
	Use:   "totp <ref>",
	Short: "Print the current TOTP code of a secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		call(http.MethodGet, "/api/secrets/"+segment(args[0])+"/totp", nil)
	},
}

var secretsFileCmd = &cobra.Command{
	Use:   "file <ref>",
	Short: "Print the content of a file secret exactly as stored",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Stdout.Write(fetch(http.MethodGet, "/api/secrets/"+segment(args[0])+"/file", nil))
	},
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Read SSH targets",
}

var sshTargetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "List stored hosts and managed machines",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		call(http.MethodGet, "/api/ssh/targets", nil)
	},
}

var machinesCmd = &cobra.Command{
	Use:   "machines",
	Short: "Read EC2 machines",
}

var machinesListCmd = &cobra.Command{
	Use:   "list <profile-ref>",
	Short: "List the managed machines of an AWS profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		call(http.MethodGet, "/api/aws/machines?profile="+url.QueryEscape(args[0]), nil)
	},
}

func init() {
	secretsCmd.AddCommand(secretsListCmd, secretsGetCmd, secretsTOTPCmd, secretsFileCmd)
	sshCmd.AddCommand(sshTargetsCmd)
	machinesCmd.AddCommand(machinesListCmd)
	ApiCmd.AddCommand(secretsCmd, sshCmd, machinesCmd)
}
