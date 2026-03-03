package cryptoCmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/internal/utils"
)

var SecretsCmd = &cobra.Command{
	Use:     "pass",
	Aliases: []string{"p"},
	Short:   "Manage secrets with AES-GCM encryption with support for single and multiline inputs and custom password",
}

var secretsFlags struct {
	secretsFile string
	multiline   bool
	password    string
	pipe        bool
	initialized bool
}

func initSecretsStore() {
	if secretsFlags.initialized {
		return
	}
	secretsFlags.initialized = true
	homeDir, err := os.UserHomeDir()
	secretsFlags.secretsFile = "secrets.json"
	if err == nil {
		anbuDir := filepath.Join(homeDir, ".config", "anbu")
		if err := os.MkdirAll(anbuDir, 0755); err != nil {
			u.PrintFatal("failed to create anbu directory", err)
		}
		secretsFlags.secretsFile = filepath.Join(anbuDir, "secrets.json")
	}
	if err := anbuCrypto.InitializeSecretsStore(secretsFlags.secretsFile); err != nil {
		u.PrintFatal("failed to initialize secrets store", err)
	}
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets with their IDs",
	Run: func(cmd *cobra.Command, args []string) {
		initSecretsStore()
		secrets, err := anbuCrypto.ListSecrets(secretsFlags.secretsFile)
		if err != nil {
			u.PrintFatal("failed to list secrets", err)
		}
		if len(secrets) == 0 {
			u.PrintInfo("No secrets found")
			return
		}
		table := u.NewTable([]string{"#", "Name"})
		for i, id := range secrets {
			table.Rows = append(table.Rows, []string{
				fmt.Sprintf("%d", i+1),
				id,
			})
		}
		table.PrintTable(false)
	},
}
var secretsGetCmd = &cobra.Command{
	Use:   "get <secret-id>",
	Short: "Print the decrypted value of a specific secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initSecretsStore()
		password := secretsFlags.password
		value, err := anbuCrypto.GetSecret(secretsFlags.secretsFile, args[0], password)
		if err != nil {
			u.PrintFatal("failed to get secret", err)
		}
		u.PrintGeneric(value)
	},
}
var secretsSetCmd = &cobra.Command{
	Use:   "add <secret-id>",
	Short: "Set the value for a secret with optional multiline input",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initSecretsStore()
		secretID := args[0]
		var value string
		if secretsFlags.pipe {
			var err error
			value, err = u.ReadPipedInput()
			if err != nil {
				u.PrintFatal("failed to read piped input", err)
			}
		} else if secretsFlags.multiline {
			value = u.GetMultilineInput(fmt.Sprintf("Enter value for secret '%s':", secretID), "")
		} else {
			value = u.GetInput(fmt.Sprintf("Enter value for secret '%s':", secretID), "")
		}
		if value == "" {
			u.PrintFatal("no value provided for secret", nil)
		}
		password := secretsFlags.password
		if err := anbuCrypto.SetSecret(secretsFlags.secretsFile, secretID, value, password); err != nil {
			u.PrintFatal("failed to set secret", err)
		}
		u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(secretID), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess("Secret set")))
	},
}
var secretsDeleteCmd = &cobra.Command{
	Use:   "delete <secret-id>",
	Short: "Delete a secret from the store",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initSecretsStore()
		if err := anbuCrypto.DeleteSecret(secretsFlags.secretsFile, args[0]); err != nil {
			u.PrintFatal("failed to delete secret", err)
		}
		u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(args[0]), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess("Secret deleted")))
	},
}
var secretsImportCmd = &cobra.Command{
	Use:   "import <file-path>",
	Short: "Import secrets from a JSON file and encrypt them in the store",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initSecretsStore()
		importFile := args[0]
		password := secretsFlags.password
		if err := anbuCrypto.ImportSecrets(secretsFlags.secretsFile, importFile, password); err != nil {
			u.PrintFatal("failed to import secrets", err)
		}
		u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(importFile), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess("Secrets imported")))
	},
}
var secretsExportCmd = &cobra.Command{
	Use:   "export <file-path>",
	Short: "Export all secrets to a JSON file in decrypted form",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initSecretsStore()
		exportFile := args[0]
		password := secretsFlags.password
		if err := anbuCrypto.ExportSecrets(secretsFlags.secretsFile, exportFile, password); err != nil {
			u.PrintFatal("failed to export secrets", err)
		}
		u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(exportFile), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess("Secrets exported")))
	},
}

func init() {
	SecretsCmd.PersistentFlags().StringVar(&secretsFlags.password, "password", "p455w0rd", "Password for encryption/decryption (default: p455w0rd)")
	secretsSetCmd.Flags().BoolVarP(&secretsFlags.multiline, "multiline", "m", false, "Enable multiline input (end with 'EOF' on a new line)")
	secretsSetCmd.Flags().BoolVar(&secretsFlags.pipe, "pipe", false, "Read secret value from piped stdin instead of interactive input")

	SecretsCmd.AddCommand(secretsListCmd)
	SecretsCmd.AddCommand(secretsGetCmd)
	SecretsCmd.AddCommand(secretsSetCmd)
	SecretsCmd.AddCommand(secretsDeleteCmd)
	SecretsCmd.AddCommand(secretsImportCmd)
	SecretsCmd.AddCommand(secretsExportCmd)
}
