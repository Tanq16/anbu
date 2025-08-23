package cryptoCmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/utils"
)

var SecretsCmd = &cobra.Command{
	Use:     "pass",
	Aliases: []string{"p"},
	Short:   "Manage secrets and parameters securely",
	Long: `Examples:
s
`,
}

var secretsFile string
var multilineFlag bool
var remoteHost string

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List IDs of all secrets and parameters",
	Run: func(cmd *cobra.Command, args []string) {
		if remoteHost != "" {
			if err := anbuCrypto.RemoteListSecrets(remoteHost); err != nil {
				u.PrintError(err.Error())
				os.Exit(1)
			}
			return
		}
		secrets, err := anbuCrypto.ListSecrets(secretsFile)
		if err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
		u.PrintSuccess("Stored secrets:")
		for i, id := range secrets {
			fmt.Printf("  %d. %s\n", i+1, u.FInfo(id))
		}
		fmt.Printf("\nTotal: %d secrets\n", len(secrets))
	},
}
var secretsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print value of a specific secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if remoteHost != "" {
			if err := anbuCrypto.RemoteGetSecret(remoteHost, args[0]); err != nil {
				u.PrintError(err.Error())
				os.Exit(1)
			}
			return
		}
		password, err := anbuCrypto.GetPassword()
		if err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
		value, err := anbuCrypto.GetSecret(secretsFile, args[0], password)
		if err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
		fmt.Println(value)
	},
}
var secretsSetCmd = &cobra.Command{
	Use:   "add",
	Short: "Set value for a secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		secretID := args[0]
		fmt.Printf("Enter value for secret '%s': ", secretID)
		var value string
		var err error
		if multilineFlag {
			value, err = anbuCrypto.ReadMultilineInput()
		} else {
			value, err = anbuCrypto.ReadSingleLineInput()
		}
		if err != nil {
			u.PrintError(fmt.Sprintf("failed to read secret value: %v", err))
			os.Exit(1)
		}
		if remoteHost != "" {
			if err := anbuCrypto.RemoteSetSecret(remoteHost, secretID, value); err != nil {
				u.PrintError(err.Error())
				os.Exit(1)
			}
			return
		}
		password, err := anbuCrypto.GetPassword()
		if err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
		if err := anbuCrypto.SetSecret(secretsFile, secretID, value, password); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
		u.PrintSuccess(fmt.Sprintf("Secret '%s' set successfully", secretID))
	},
}
var secretsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if remoteHost != "" {
			if err := anbuCrypto.RemoteDeleteSecret(remoteHost, args[0]); err != nil {
				u.PrintError(err.Error())
				os.Exit(1)
			}
			return
		}
		if err := anbuCrypto.DeleteSecret(secretsFile, args[0]); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
		u.PrintSuccess(fmt.Sprintf("Secret '%s' deleted successfully", args[0]))
	},
}
var secretsImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import secrets and parameters from file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importFile := args[0]
		if remoteHost != "" {
			if err := anbuCrypto.RemoteImportSecrets(remoteHost, importFile); err != nil {
				u.PrintError(err.Error())
				os.Exit(1)
			}
			return
		}
		if err := anbuCrypto.ImportSecrets(secretsFile, importFile); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
		u.PrintSuccess(fmt.Sprintf("Imported secrets from %s successfully", importFile))
	},
}
var secretsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export secrets and parameters to file (unencrypted)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exportFile := args[0]
		if remoteHost != "" {
			if err := anbuCrypto.RemoteExportSecrets(remoteHost, exportFile); err != nil {
				u.PrintError(err.Error())
				os.Exit(1)
			}
			return
		}
		if err := anbuCrypto.ExportSecrets(secretsFile, exportFile); err != nil {
			u.PrintError(err.Error())
			os.Exit(1)
		}
		u.PrintSuccess(fmt.Sprintf("Exported secrets to %s successfully", exportFile))
	},
}

var secretsServe = &cobra.Command{
	Use:   "serve",
	Short: "Use a server to handle secrets on a remote server via a simple web API",
	Long: `Example:
Start the server with: a p serve
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := secretsFile
		if len(args) > 0 {
			filePath = args[0]
		}
		if err := anbuCrypto.ServeSecrets(filePath); err != nil {
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

	SecretsCmd.PersistentFlags().StringVar(&remoteHost, "remote", "", "Remote server URL to make API calls to")
	secretsSetCmd.Flags().BoolVarP(&multilineFlag, "multiline", "m", false, "Enable multiline input (end with `EOF` on a new line)")

	SecretsCmd.AddCommand(secretsListCmd)
	SecretsCmd.AddCommand(secretsGetCmd)
	SecretsCmd.AddCommand(secretsSetCmd)
	SecretsCmd.AddCommand(secretsDeleteCmd)
	SecretsCmd.AddCommand(secretsImportCmd)
	SecretsCmd.AddCommand(secretsExportCmd)
	SecretsCmd.AddCommand(secretsServe)
}
