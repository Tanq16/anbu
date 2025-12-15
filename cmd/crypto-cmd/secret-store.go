package cryptoCmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/utils"
)

var SecretsCmd = &cobra.Command{
	Use:     "pass",
	Aliases: []string{"p"},
	Short:   "Manage secrets securely with AES-GCM encryption",
	Long: `A secure store for secrets, which are encrypted at rest using AES-GCM.
The store uses a default password of "p455w0rd" unless a custom password is
specified with the --password flag.

Examples:
  # List all stored secret IDs
  anbu pass list

  # Add a new secret
  anbu pass add my-api-key

  # Add a multi-line secret like a private key
  anbu pass add my-ssh-key -m

  # Retrieve and print a secret's value
  anbu pass get my-api-key

  # Delete a secret
  anbu pass delete my-api-key

  # Export all secrets (decrypted) to a JSON file
  anbu pass export secrets_backup.json

  # Import secrets from a JSON file
  anbu pass import secrets_backup.json`,
}

var secretsFile string
var multilineFlag bool
var passwordFlag string

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets",
	Run: func(cmd *cobra.Command, args []string) {
		secrets, err := anbuCrypto.ListSecrets(secretsFile)
		if err != nil {
			log.Fatal().Err(err)
		}
		u.PrintSuccess("Stored secrets:")
		for i, id := range secrets {
			fmt.Printf("  %d. %s\n", i+1, u.FInfo(id))
		}
		fmt.Printf("\nTotal: %d secrets\n", len(secrets))
	},
}
var secretsGetCmd = &cobra.Command{
	Use:   "get <secret-id>",
	Short: "Print the value of a specific secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		password := passwordFlag
		value, err := anbuCrypto.GetSecret(secretsFile, args[0], password)
		if err != nil {
			log.Fatal().Err(err)
		}
		fmt.Println(value)
	},
}
var secretsSetCmd = &cobra.Command{
	Use:   "add <secret-id>",
	Short: "Set the value for a secret",
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
			log.Fatal().Err(err).Msg("failed to read secret value")
		}
		password := passwordFlag
		if err := anbuCrypto.SetSecret(secretsFile, secretID, value, password); err != nil {
			log.Fatal().Err(err)
		}
		log.Info().Msgf("Secret '%s' set successfully", secretID)
	},
}
var secretsDeleteCmd = &cobra.Command{
	Use:   "delete <secret-id>",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCrypto.DeleteSecret(secretsFile, args[0]); err != nil {
			log.Fatal().Err(err)
		}
		log.Info().Msgf("Secret '%s' deleted successfully", args[0])
	},
}
var secretsImportCmd = &cobra.Command{
	Use:   "import <file-path>",
	Short: "Import secrets from a JSON file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importFile := args[0]
		password := passwordFlag
		if err := anbuCrypto.ImportSecrets(secretsFile, importFile, password); err != nil {
			log.Fatal().Err(err)
		}
		u.PrintSuccess(fmt.Sprintf("Imported secrets from %s successfully", importFile))
	},
}
var secretsExportCmd = &cobra.Command{
	Use:   "export <file-path>",
	Short: "Export secrets to a JSON file (unencrypted)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exportFile := args[0]
		password := passwordFlag
		if err := anbuCrypto.ExportSecrets(secretsFile, exportFile, password); err != nil {
			log.Fatal().Err(err)
		}
		log.Info().Msgf("Exported secrets to %s successfully", exportFile)
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
		log.Fatal().Err(err)
	}

	// default known password is fine here - we would anyway use a password inline or in env. var. direct workstation access is anyway a risk, intention is secrets are not in plain text in history or logs. not a huge risk with default password, but option for custom password is nice.
	SecretsCmd.Flags().StringVar(&passwordFlag, "password", "p455w0rd", "Password for encryption/decryption (default: p455w0rd)")
	secretsSetCmd.Flags().BoolVarP(&multilineFlag, "multiline", "m", false, "Enable multiline input (end with 'EOF' on a new line)")

	SecretsCmd.AddCommand(secretsListCmd)
	SecretsCmd.AddCommand(secretsGetCmd)
	SecretsCmd.AddCommand(secretsSetCmd)
	SecretsCmd.AddCommand(secretsDeleteCmd)
	SecretsCmd.AddCommand(secretsImportCmd)
	SecretsCmd.AddCommand(secretsExportCmd)
}
