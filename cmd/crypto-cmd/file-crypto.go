package cryptoCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
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

var FileCryptoCmd = &cobra.Command{
	Use:   "file-crypt",
	Short: "Encryption/decryption on files using AES-256-GCM symmetric encryption",
}

var fileCryptoEncryptSymmCmd = &cobra.Command{
	Use:   "encrypt",
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
		err := anbuCrypto.EncryptFileSymmetric(fileCryptoFlags.file, fileCryptoFlags.password)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to encrypt file")
		}
		fmt.Println(utils.OutDetail("File encrypted successfully: ") + utils.OutSuccess(fileCryptoFlags.file+".enc"))
	},
}

var fileCryptoDecryptSymmCmd = &cobra.Command{
	Use:   "decrypt",
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
		err := anbuCrypto.DecryptFileSymmetric(fileCryptoFlags.file, fileCryptoFlags.password)
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

func init() {
	FileCryptoCmd.AddCommand(fileCryptoEncryptSymmCmd)
	FileCryptoCmd.AddCommand(fileCryptoDecryptSymmCmd)

	fileCryptoEncryptSymmCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Encryption password")
	fileCryptoDecryptSymmCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Decryption password")
}
