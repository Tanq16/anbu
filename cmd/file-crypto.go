package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuFileCrypto "github.com/tanq16/anbu/internal/filecrypto"
	"github.com/tanq16/anbu/utils"
)

var fileCryptoFlags struct {
	file     string
	password string
}

var fileCryptoCmd = &cobra.Command{
	Use:   "filecrypto",
	Short: "Encryption operations on a file",
}

var fileCryptoEncryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt a file using AES-256-ECB",
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
		err := anbuFileCrypto.EncryptFile(fileCryptoFlags.file, fileCryptoFlags.password)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to encrypt file")
		}
		fmt.Println(utils.OutDetail("File encrypted successfully: ") + utils.OutSuccess(fileCryptoFlags.file+".enc"))
	},
}

var fileCryptoDecryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt a file that was encrypted using AES-256-ECB",
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
		err := anbuFileCrypto.DecryptFile(fileCryptoFlags.file, fileCryptoFlags.password)
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
	rootCmd.AddCommand(fileCryptoCmd)
	fileCryptoCmd.AddCommand(fileCryptoEncryptCmd)
	fileCryptoCmd.AddCommand(fileCryptoDecryptCmd)

	fileCryptoEncryptCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Encryption password")
	fileCryptoDecryptCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Decryption password")
}
