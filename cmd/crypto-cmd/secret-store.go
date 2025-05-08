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
	Long: `Manage secrets and parameters securely. Secrets are stored encrypted, while parameters are stored in plain text.
Examples:
  anbu secrets list                  # List all secrets and parameters
  anbu secrets get SECRETID          # Show specific secret
  anbu secrets param get PARAMID     # Show specific parameter
  anbu secrets set SECRETID          # Set secret (prompt for value)
  anbu secrets delete SECRETID       # Delete secret
  anbu secrets param set PARAMID     # Set parameter (prompt for value)
  anbu secrets param delete PARAMID  # Delete parameter
  anbu secrets import FILE           # Import secrets and parameters from file
  anbu secrets export FILE           # Export secrets and parameters to file (unencrypted)`,
}

var secretsFile string

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets and parameters",
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.ListSecrets(secretsFile); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}

var secretsGetCmd = &cobra.Command{
	Use:   "get SECRETID",
	Short: "Show a specific secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.GetSecret(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}

var secretsSetCmd = &cobra.Command{
	Use:   "set SECRETID",
	Short: "Set a secret (prompt for value)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.SetSecret(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}

var secretsDeleteCmd = &cobra.Command{
	Use:   "delete SECRETID",
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
	Use:   "import FILE",
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
	Use:   "export FILE",
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
	Use:   "param",
	Short: "Manage parameters",
}

var secretsParamGetCmd = &cobra.Command{
	Use:   "get PARAMID",
	Short: "Show a specific parameter",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.GetParameter(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}

var secretsParamSetCmd = &cobra.Command{
	Use:   "set PARAMID",
	Short: "Set a parameter (prompt for value)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.SetParameter(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
	},
}

var secretsParamDeleteCmd = &cobra.Command{
	Use:   "delete PARAMID",
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
	// Get home directory for default file path
	homeDir, err := os.UserHomeDir()
	defaultPath := ""
	if err == nil {
		defaultPath = filepath.Join(homeDir, ".anbu-secrets.json")
	}

	// Add global flag for secrets file location
	SecretsCmd.PersistentFlags().StringVarP(&secretsFile, "file", "f", defaultPath, "Path to secrets file")

	// Add commands
	SecretsCmd.AddCommand(secretsListCmd)
	SecretsCmd.AddCommand(secretsGetCmd)
	SecretsCmd.AddCommand(secretsSetCmd)
	SecretsCmd.AddCommand(secretsDeleteCmd)
	SecretsCmd.AddCommand(secretsImportCmd)
	SecretsCmd.AddCommand(secretsExportCmd)

	// Add parameter subcommands
	secretsParamCmd.AddCommand(secretsParamGetCmd)
	secretsParamCmd.AddCommand(secretsParamSetCmd)
	secretsParamCmd.AddCommand(secretsParamDeleteCmd)
	SecretsCmd.AddCommand(secretsParamCmd)
}
