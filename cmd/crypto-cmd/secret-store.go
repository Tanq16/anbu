package cryptoCmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/utils"
)

var SecretsCmd = &cobra.Command{
	Use:     "secrets",
	Aliases: []string{"p"},
	Short:   "Manage secrets with AES-GCM encryption with support for single and multiline inputs and custom password",
}

var secretsFlags struct {
	secretsFile string
	multiline   bool
	password    string
	value       string
	valueFile   string
	filter      string
	initialized bool
}

func initSecretsStore() {
	if secretsFlags.initialized {
		return
	}
	secretsFlags.initialized = true
	secretsFlags.secretsFile = filepath.Join(u.ConfigDir(), "secrets.json")
	if err := anbuCrypto.InitializeSecretsStore(secretsFlags.secretsFile); err != nil {
		u.PrintFatal("failed to initialize secrets store", err)
	}
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets with their IDs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		initSecretsStore()
		secrets, err := anbuCrypto.ListSecrets(secretsFlags.secretsFile)
		if err != nil {
			u.PrintFatal("failed to list secrets", err)
		}
		if secretsFlags.filter != "" {
			re, err := regexp.Compile(secretsFlags.filter)
			if err != nil {
				u.PrintFatal("invalid --filter regex", err)
			}
			secrets = slices.DeleteFunc(secrets, func(name string) bool {
				return !re.MatchString(name)
			})
		}
		if len(secrets) == 0 {
			u.PrintInfo("No secrets found")
			return
		}
		width := len(strconv.Itoa(len(secrets)))
		for i, name := range secrets {
			u.PrintGeneric(fmt.Sprintf(" %*d. %s", width, i+1, name))
		}
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
		value := secretsFlags.value
		if secretsFlags.valueFile != "" {
			loaded, err := u.ReadFileFlag(cmd, "value-file")
			if err != nil {
				u.PrintFatal("failed to read --value-file", err)
			}
			value = loaded
		}
		if value == "" {
			var err error
			if secretsFlags.multiline {
				value, err = u.PromptTextArea(fmt.Sprintf("Enter value for secret '%s':", secretID), "")
			} else {
				value, err = u.PromptInput(fmt.Sprintf("Enter value for secret '%s':", secretID), "")
			}
			if errors.Is(err, u.ErrNoTerminal) {
				u.PrintFatal("add needs --value, --value -, --value-file, or --value-file -", nil)
			}
			if err != nil {
				u.PrintFatal("failed to read secret value", err)
			}
		}
		if value == "" {
			u.PrintFatal("add needs --value, --value -, --value-file, or --value-file -", nil)
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
	passwordHelp := "Password for encryption/decryption (default: p455w0rd)"
	secretsGetCmd.Flags().StringVar(&secretsFlags.password, "password", "p455w0rd", passwordHelp)
	secretsSetCmd.Flags().StringVar(&secretsFlags.password, "password", "p455w0rd", passwordHelp)
	secretsImportCmd.Flags().StringVar(&secretsFlags.password, "password", "p455w0rd", passwordHelp)
	secretsExportCmd.Flags().StringVar(&secretsFlags.password, "password", "p455w0rd", passwordHelp)
	secretsListCmd.Flags().StringVarP(&secretsFlags.filter, "filter", "f", "", "Regex; keep names that match")
	secretsSetCmd.Flags().StringVar(&secretsFlags.value, "value", "", "Secret value, or - to read it from stdin")
	secretsSetCmd.Flags().StringVar(&secretsFlags.valueFile, "value-file", "", "File containing the secret value, or - for stdin")
	secretsSetCmd.Flags().BoolVar(&secretsFlags.multiline, "multiline", false, "Prompt with a multi-line editor when no value flag is set")
	secretsSetCmd.MarkFlagsMutuallyExclusive("value", "value-file")
	_ = u.MarkStdinLine(secretsSetCmd, "value")
	_ = u.MarkStdinStream(secretsSetCmd, "value-file")
	SecretsCmd.AddCommand(secretsListCmd)
	SecretsCmd.AddCommand(secretsGetCmd)
	SecretsCmd.AddCommand(secretsSetCmd)
	SecretsCmd.AddCommand(secretsDeleteCmd)
	SecretsCmd.AddCommand(secretsImportCmd)
	SecretsCmd.AddCommand(secretsExportCmd)
}
