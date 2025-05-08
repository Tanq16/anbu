package cryptoCmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/utils"
)

var SecretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets and parameters securely",
}

var secretsFile string

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List IDs of all secrets and parameters",
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.ListSecrets(secretsFile); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}
var secretsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print value of a specific secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.GetSecret(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}
var secretsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set value for a secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.SetSecret(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}
var secretsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.DeleteSecret(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}
var secretsImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import secrets and parameters from file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importFile := args[0]
		if err := anbuCrypto.ImportSecrets(secretsFile, importFile); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}
var secretsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export secrets and parameters to file (unencrypted)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exportFile := args[0]
		if err := anbuCrypto.ExportSecrets(secretsFile, exportFile); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}

// Parameter commands
var secretsParamCmd = &cobra.Command{
	Use:   "p",
	Short: "Manage parameters",
}
var secretsParamGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print value of a specific parameter",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.GetParameter(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}
var secretsParamSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set value for a parameter",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.SetParameter(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}
var secretsParamDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a parameter",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.DeleteParameter(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	secretsFile = ".anbu-secrets.json"
	if err == nil {
		secretsFile = filepath.Join(homeDir, ".anbu-secrets.json")
	}
	err = anbuCrypto.InitializeSecretsStore(secretsFile)
	if err != nil {
		u.PrintError(err.Error())
		os.Exit(1)
	}

	SecretsCmd.AddCommand(secretsListCmd)
	SecretsCmd.AddCommand(secretsGetCmd)
	SecretsCmd.AddCommand(secretsSetCmd)
	SecretsCmd.AddCommand(secretsDeleteCmd)
	SecretsCmd.AddCommand(secretsImportCmd)
	SecretsCmd.AddCommand(secretsExportCmd)

	SecretsCmd.AddCommand(secretsParamCmd)

	secretsParamCmd.AddCommand(secretsParamGetCmd)
	secretsParamCmd.AddCommand(secretsParamSetCmd)
	secretsParamCmd.AddCommand(secretsParamDeleteCmd)
}
