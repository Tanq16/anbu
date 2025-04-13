package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuFileCrypto "github.com/tanq16/anbu/internal/filecrypto"
	"github.com/tanq16/anbu/utils"
)

var fileCryptoFlags struct {
	file          string
	password      string
	pubKeyPath    string
	privKeyPath   string
	signerPubKey  string
	signerPrivKey string
	passphrase    string
}

var fileCryptoCmd = &cobra.Command{
	Use:   "filecrypto",
	Short: "Encryption operations on files",
}

var fileCryptoEncryptSymmCmd = &cobra.Command{
	Use:   "encrypt-symm",
	Short: "Encrypt a file using AES-256-GCM symmetric encryption",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("filecrypto")
		fileCryptoFlags.file = args[0]
		if fileCryptoFlags.file == "" {
			logger.Fatal().Msg("No input file specified")
		}
		if fileCryptoFlags.password == "" {
			logger.Fatal().Msg("No password specified")
		}
		err := anbuFileCrypto.EncryptSymmetric(fileCryptoFlags.file, fileCryptoFlags.password)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to encrypt file")
		}
		fmt.Println(utils.OutDetail("File encrypted successfully: ") + utils.OutSuccess(fileCryptoFlags.file+".enc"))
	},
}

var fileCryptoDecryptSymmCmd = &cobra.Command{
	Use:   "decrypt-symm",
	Short: "Decrypt a file that was encrypted using AES-256-GCM symmetric encryption",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("filecrypto")
		fileCryptoFlags.file = args[0]
		if fileCryptoFlags.file == "" {
			logger.Fatal().Msg("No input file specified")
		}
		if fileCryptoFlags.password == "" {
			logger.Fatal().Msg("No password specified")
		}
		err := anbuFileCrypto.DecryptSymmetric(fileCryptoFlags.file, fileCryptoFlags.password)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to decrypt file")
		}
		outputFile := fileCryptoFlags.file
		if len(outputFile) > 4 && outputFile[len(outputFile)-4:] == ".enc" {
			outputFile = outputFile[:len(outputFile)-4]
		}
		fmt.Println(utils.OutDetail("File decrypted successfully: ") + utils.OutSuccess(outputFile))
	},
}

var fileCryptoEncryptPGPCmd = &cobra.Command{
	Use:   "encrypt-pgp",
	Short: "Encrypt a file using PGP-like encryption (RSA + AES-GCM)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("filecrypto")
		fileCryptoFlags.file = args[0]
		if fileCryptoFlags.file == "" {
			logger.Fatal().Msg("No input file specified")
		}
		if fileCryptoFlags.pubKeyPath == "" {
			logger.Fatal().Msg("No recipient public key specified")
		}
		err := anbuFileCrypto.EncryptPGP(
			fileCryptoFlags.file,
			fileCryptoFlags.pubKeyPath,
			fileCryptoFlags.signerPrivKey,
			fileCryptoFlags.passphrase,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to PGP encrypt file")
		}
		fmt.Println(utils.OutDetail("File PGP encrypted successfully: ") + utils.OutSuccess(fileCryptoFlags.file+".pgp"))
		if fileCryptoFlags.signerPrivKey != "" {
			fmt.Println(utils.OutInfo("File was signed with the provided private key"))
		}
	},
}

var fileCryptoDecryptPGPCmd = &cobra.Command{
	Use:   "decrypt-pgp",
	Short: "Decrypt a file that was encrypted using PGP-like encryption",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("filecrypto")
		fileCryptoFlags.file = args[0]
		if fileCryptoFlags.file == "" {
			logger.Fatal().Msg("No input file specified")
		}
		if fileCryptoFlags.privKeyPath == "" {
			logger.Fatal().Msg("No recipient private key specified")
		}
		err := anbuFileCrypto.DecryptPGP(
			fileCryptoFlags.file,
			fileCryptoFlags.privKeyPath,
			fileCryptoFlags.signerPubKey,
			fileCryptoFlags.passphrase,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to PGP decrypt file")
		}
		outputFile := fileCryptoFlags.file
		if len(outputFile) > 4 && outputFile[len(outputFile)-4:] == ".pgp" {
			outputFile = outputFile[:len(outputFile)-4]
		}
		fmt.Println(utils.OutDetail("File PGP decrypted successfully: ") + utils.OutSuccess(outputFile))
		if fileCryptoFlags.signerPubKey != "" {
			fmt.Println(utils.OutInfo("Signature verification was performed"))
		}
	},
}

func init() {
	rootCmd.AddCommand(fileCryptoCmd)

	// Add symmetric encryption commands
	fileCryptoCmd.AddCommand(fileCryptoEncryptSymmCmd)
	fileCryptoCmd.AddCommand(fileCryptoDecryptSymmCmd)

	// Add PGP encryption commands
	fileCryptoCmd.AddCommand(fileCryptoEncryptPGPCmd)
	fileCryptoCmd.AddCommand(fileCryptoDecryptPGPCmd)

	// Symmetric encryption flags
	fileCryptoEncryptSymmCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Encryption password")
	fileCryptoDecryptSymmCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Decryption password")

	// PGP encryption flags
	fileCryptoEncryptPGPCmd.Flags().StringVarP(&fileCryptoFlags.pubKeyPath, "recipient-key", "r", "", "Recipient's public key path")
	fileCryptoEncryptPGPCmd.Flags().StringVarP(&fileCryptoFlags.signerPrivKey, "signer-key", "s", "", "Signer's private key path (optional)")
	fileCryptoEncryptPGPCmd.Flags().StringVarP(&fileCryptoFlags.passphrase, "passphrase", "p", "", "Passphrase for signer's private key")

	// PGP decryption flags
	fileCryptoDecryptPGPCmd.Flags().StringVarP(&fileCryptoFlags.privKeyPath, "recipient-key", "r", "", "Recipient's private key path")
	fileCryptoDecryptPGPCmd.Flags().StringVarP(&fileCryptoFlags.signerPubKey, "signer-key", "s", "", "Signer's public key path for verification (optional)")
	fileCryptoDecryptPGPCmd.Flags().StringVarP(&fileCryptoFlags.passphrase, "passphrase", "p", "", "Passphrase for recipient's private key")

	// Mark required flags
	fileCryptoEncryptPGPCmd.MarkFlagRequired("recipient-key")
	fileCryptoDecryptPGPCmd.MarkFlagRequired("recipient-key")
}
