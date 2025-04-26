package cryptoCmd

import (
	"os"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/utils"
)

var fileCryptoFlags struct {
	file     string
	password string
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
		if fileCryptoFlags.password == "" {
			u.PrintError("No password specified")
			os.Exit(1)
		}
		anbuCrypto.EncryptFileSymmetric(args[0], fileCryptoFlags.password)
	},
}

var fileCryptoDecryptSymmCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt a file that was encrypted using AES-256-GCM symmetric encryption",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if fileCryptoFlags.password == "" {
			u.PrintError("No password specified")
			os.Exit(1)
		}
		anbuCrypto.DecryptFileSymmetric(args[0], fileCryptoFlags.password)
	},
}

func init() {
	FileCryptoCmd.AddCommand(fileCryptoEncryptSymmCmd)
	FileCryptoCmd.AddCommand(fileCryptoDecryptSymmCmd)

	fileCryptoEncryptSymmCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Encryption password")
	fileCryptoDecryptSymmCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Decryption password")
}
